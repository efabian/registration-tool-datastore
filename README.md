# Registration Tool

An AppEngine-based project for simple registration.
Can also be used outsite GCP platform if you decouple the Datastore and use another database.

## Required Development Environment

Install `GCloud SDK` and the following components: 
 - cloud-datastore-emulator
 - app-engine-go

## Local Deployment

### With Runtime Go111 or newer

In the new Datastore package, the client has been decoupled from the App Engine, so a service account has to be created.
Also, in Go111, the project's datastore will be the one to use.
- Create a new service account
- Set the account as project owner
- Create a key for the service account in JSON
- Export the key in your OS environment: `export GOOGLE_APPLICATION_CREDENTIALS="[PATH]"`
- Management endpoint will be: `http://localhost:8000`
- API enpoint will be: `http://localhost:8080`


### With Runtime Go1
This will start a datastore emulator

Management endpoint will be: `http://localhost:8000`
API enpoint will be: `http://localhost:8080`

```
$dev_appserver.py app.yaml
```

## Production Deployment

1. Set GCloud project to `${PROJECT_NAME}` 
2. Make sure that the project is indeed set to `${PROJECT_NAME}`
3. Deploy the project's app.yaml

```
$gcloud config set project ${PROJECT_NAME}
$gcloud config list
$gcloud app deploy app.yaml
```

## API endpoints and behaviour

### https://${DOMAIN}/internal/register
- Main registration link

### https://${DOMAIN}/internal/retrieve
- To retrieve data in the database #ToDo


