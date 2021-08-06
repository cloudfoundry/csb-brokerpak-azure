import express from 'express'
import helmet from 'helmet'
import vcapServices from 'vcap_services'
import { BlobServiceClient, StorageSharedKeyCredential } from '@azure/storage-blob'

const port = process.env.PORT || 8080

const main = async () => {
  console.log('starting')

  const credentials = vcapServices.findCredentials({ instance: { tags: 'storage' } })
  if (typeof credentials !== 'object' || Object.entries(credentials).length === 0) {
    throw new Error('could not find credentials in VCAP_SERVICES')
  }

  console.log('connecting to storage account')
  const sharedKeyCredential = new StorageSharedKeyCredential(credentials.storage_account_name, credentials.primary_access_key)
  const client = new BlobServiceClient(
    `https://${credentials.storage_account_name}.blob.core.windows.net`,
    sharedKeyCredential,
  )

  const app = express()
  app.use(helmet())
  app.use(express.text({limit: '1kb', type: '*/*'}))
  app.get('/', handleListContainers(client))
  app.put('/:container', handleCreateContainer(client))
  app.put('/:container/:blobname', handleCreateBlob(client))
  app.get('/:container/:blobname', handleFetchBlob(client))

  app.listen(port, () => console.log(`listening on port ${port}`))
}

const handleListContainers = (client) => async (req, res) => {
  try {
    console.log('handling list containers request')
    let list = []
    const iter = await client.listContainers()
    for await (const container of iter) {
      list.push(container.name)
    }

    console.log('result: ' + JSON.stringify(list))
    res.json(list)
  } catch (e) {
    res.status(500).send(e)
  }
}

const handleCreateContainer = (client) => async (req, res) => {
  try {
    const container = req.params.container

    if (typeof container !== 'string' || container.length === 0) {
      console.log('container name not specified', req.body)
      res.status(400).send('container name not specified')
      return
    }

    console.log(`handling create container "${container}".`)
    const containerClient = client.getContainerClient(container)
    const result = await containerClient.create()
    console.log(`container ${container} created successfully`, result.requestId)

    res.sendStatus(200)
  } catch (e) {
    console.log('caught', e)
    res.status(500).send(e)
  }
}

const handleCreateBlob = (client) => async (req, res) => {
  try {
    const container = req.params.container
    const blobname = req.params.blobname
    const blobdata = req.body

    if (typeof container !== 'string' || container.length === 0) {
      console.log('container name not specified', req.body)
      res.status(400).send('container name not specified')
      return
    }

    if (typeof blobname !== 'string' || blobname.length === 0) {
      console.log('blob name not specified', req.body)
      res.status(400).send('blob name not specified')
      return
    }

    if (typeof blobdata !== 'string' || blobdata.length === 0) {
      console.log('blob data not specified', blobdata)
      res.status(400).send('blob data not specified')
      return
    }

    const containerClient = client.getContainerClient(container)
    const blockBlobClient = containerClient.getBlockBlobClient(blobname)
    const result = await blockBlobClient.upload(blobdata, blobdata.length)
    console.log(`uploaded blob ${blobname} successfully`, result.requestId);

    res.sendStatus(200)
  } catch (e) {
    console.log('caught', e)
    res.status(500).send(e)
  }
}

const handleFetchBlob = (client) => async (req, res) => {
  try {
    const container = req.params.container
    const blobname = req.params.blobname

    if (typeof container !== 'string' || container.length === 0) {
      console.log('container name not specified', req.body)
      res.status(400).send('container name not specified')
      return
    }

    if (typeof blobname !== 'string' || blobname.length === 0) {
      console.log('blob name not specified', req.body)
      res.status(400).send('blob name not specified')
      return
    }

    console.log(`handling fetch blob ${blobname} in container ${container}`)
    const containerClient = client.getContainerClient(container)
    const blockBlobClient = containerClient.getBlockBlobClient(blobname)
    const downloadResponse = await blockBlobClient.download()
    const data = (await streamToBuffer(downloadResponse.readableStreamBody)).toString()
    console.log(`result: ${data}`)
    res.send(data)
  } catch (e) {
    res.status(500).send(e)
  }
}

const streamToBuffer = async (readableStream) => {
  return new Promise((resolve, reject) => {
    const chunks = []
    readableStream.on("data", (data) => {
      chunks.push(data instanceof Buffer ? data : Buffer.from(data))
    })
    readableStream.on("end", () => {
      resolve(Buffer.concat(chunks))
    })
    readableStream.on("error", reject)
  })
}

(async () => {
  try {
    await main()
  } catch (e) {
    console.error(`failed: ${e}`)
  }
})()
