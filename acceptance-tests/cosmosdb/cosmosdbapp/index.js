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
  app.use(express.json())
  app.get('/', handleListDatabases(client))
  app.get('/:database', handleListContainers(client))
  app.post('/:database', handleCreateContainer(client))
  app.post('/:database/:container', handleCreateDocument(client))
  app.get('/:database/:container/:name', handleFetchDocument(client))

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
    const container = req.body.id

    if (typeof container !== 'string' || container.length === 0) {
      console.log('container name not specified', req.body)
      res.status(400).send('container name not specified - needs a JSON object with key: id')
      return
    }

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
    const name = req.body.name
    const data = req.body.data

    if (typeof name !== 'string' || container.length === 0) {
      console.log('name name not specified', req.body)
      res.status(400).send('name name not specified - needs a JSON object with key: name')
      return
    }

    if (typeof data !== 'string' || container.length === 0) {
      console.log('data name not specified', req.body)
      res.status(400).send('data name not specified - needs a JSON object with key: data')
      return
    }

    console.log(`handling create document "${name}" with data "${data}" in container "${container}" of database ${database}`)
    const result = await client.database(database).container(container).items.create({ name, data })
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
    const name = req.params.name

    console.log(`handling fetch document request for ${name} in container ${container} for database ${database}`)
    const result = await client.database(database).container(container).items.readAll().fetchAll()
    const data = jsonata(`resources[name="${name}"].data`).evaluate(result)
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
