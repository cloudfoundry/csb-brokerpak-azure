import express from 'express'
import helmet from 'helmet'
import vcapServices from 'vcap_services'
import jsonata from 'jsonata'
import { CosmosClient } from '@azure/cosmos'

const port = process.env.PORT || 8080

const main = async () => {
  console.log('starting')

  const credentials = vcapServices.findCredentials({ instance: { tags: 'cosmosdb' } })
  if (typeof credentials !== 'object' || Object.entries(credentials).length === 0) {
    throw new Error('could not find credentials in VCAP_SERVICES')
  }

  console.log('connecting to Cosmos DB')
  const client = new CosmosClient({ endpoint: credentials.cosmosdb_host_endpoint, key: credentials.cosmosdb_master_key })

  const app = express()
  app.use(helmet())
  app.use(express.text({ limit: '1kb', type: '*/*' }))
  app.get('/', handleListDatabases(client))
  app.get('/:database', handleListContainers(client))
  app.put('/:database/:container', handleCreateContainer(client))
  app.put('/:database/:container/:document', handleCreateDocument(client))
  app.get('/:database/:container/:document', handleFetchDocument(client))

  app.listen(port, () => console.log(`listening on port ${port}`))
}

const handleListDatabases = (client) => async (req, res) => {
  try {
    console.log('handling list databases request')
    const result = await client.databases.readAll().fetchAll()
    const list = jsonata('resources.id[]').evaluate(result)
    console.log('result: ' + JSON.stringify(list))
    res.json(list)
  } catch (e) {
    res.status(500).send(e)
  }
}

const handleListContainers = (client) => async (req, res) => {
  try {
    const database = req.params.database
    console.log(`handling list containers request on database: ${database}`)
    const result = await client.database(database).containers.readAll().fetchAll()
    const list = jsonata('resources.id[]').evaluate(result)
    console.log('result: ' + JSON.stringify(list))
    res.json(list)
  } catch (e) {
    res.status(500).send(e)
  }
}

const handleCreateContainer = (client) => async (req, res) => {
  try {
    const database = req.params.database
    const container = req.params.container

    console.log(`handling create container "${container}" in database ${database}`)
    const result = await client.database(database).containers.createIfNotExists({ id: container })
    if (result.statusCode !== 201) {
      console.log('failed to create container', result)
      res.status(401).send(`failed to create container - status code ${result.statusCode}`)
      return
    }

    res.sendStatus(200)
  } catch (e) {
    console.log('caught', e)
    res.status(500).send(e)
  }
}

const handleCreateDocument = (client) => async (req, res) => {
  try {
    const database = req.params.database
    const container = req.params.container
    const document = req.params.document
    const data = req.body

    if (typeof data !== 'string' || data.length === 0) {
      console.log('no data specified', data)
      res.status(400).send('no data specified')
      return
    }

    console.log(`handling create document "${document}" with data "${data}" in container "${container}" of database ${database}`)
    const result = await client.database(database).container(container).items.create({ name: document, data })
    if (result.statusCode !== 201) {
      console.log('failed to create document', result)
      res.status(401).send(`failed to create document - status code ${result.statusCode}`)
      return
    }

    res.sendStatus(200)
  } catch (e) {
    console.log('caught', e)
    res.status(500).send(e)
  }
}

const handleFetchDocument = (client) => async (req, res) => {
  try {
    const database = req.params.database
    const container = req.params.container
    const document = req.params.document

    console.log(`handling fetch document request for ${document} in container ${container} for database ${database}`)
    const result = await client.database(database).container(container).items.readAll().fetchAll()
    const data = jsonata(`resources[name="${document}"].data`).evaluate(result)
    console.log(`result: ${data}`)
    res.send(data)
  } catch (e) {
    res.status(500).send(e)
  }
}

(async () => {
  try {
    await main()
  } catch (e) {
    console.error(`failed: ${e}`)
  }
})()
