This repo contains source code for a GCP Cloud Function. It listens for log
lines over Pub/Sub, and forwards them to Mezmo (previously called LogDNA).

The Cloud Function currently expects log lines from Cloud Run services. No idea
how it fares on log lines from other GCP sources.

You need to set up a logging sink that exports to a Pub/Sub topic. Docs on this
can be found
[here](https://cloud.google.com/logging/docs/export/configure_export_v2).

## Deploy:

```bash
$ INGESTION_KEY=<LOGDNA ingestion key>
$ PUBSUB_LOG_TOPIC=<Pub/Sub log topic name>
$ REGION=<Desired GCP Region>
$ gcloud functions deploy logdna-gcp \
      --runtime=go120 \
      --gen2 \
      --region=$REGION \
      --source=. \
      --entry-point=LogDNAUpload \
      --set-env-vars=INGESTION_KEY=$INGESTION_KEY \
      --trigger-topic=$PUBSUB_LOG_TOPIC
```

## Run locally:

```bash
$ make dev
```

## Call locally:

See [docs](https://cloud.google.com/functions/docs/running/calling).

```bash
# Assumes you have a file called data.json
$ set data (cat data.json | jq --compact-output | base64 --wrap 0)
$ curl localhost:8080 \
  -X POST \
  -H "Content-Type: application/json" \
  -H "ce-id: 123451234512345" \
  -H "ce-specversion: 1.0" \
  -H "ce-time: 2020-01-02T12:34:56.789Z" \
  -H "ce-type: google.cloud.pubsub.topic.v1.messagePublished" \
  -H "ce-source: //pubsub.googleapis.com/projects/MY-PROJECT/topics/MY-TOPIC" \
  -d '{
        "message": {
          "data": "'$data'",
        },
        "subscription": "projects/MY-PROJECT/subscriptions/MY-SUB"
      }'

```
