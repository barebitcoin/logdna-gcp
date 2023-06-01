dev:
	go build -o logdna-gcp ./cmd
	@echo serving on localhost:8080
	env FUNCTION_TARGET=LogDNAUpload ./logdna-gcp

deploy: 
	gcloud functions deploy logdna-gcp \
      --runtime=go120 \
      --gen2 \
      --region=${REGION} \
      --source=. \
      --entry-point=LogDNAUpload \
      --set-env-vars=INGESTION_KEY=${INGESTION_KEY} \
      --trigger-topic=${PUBSUB_LOG_TOPIC}
