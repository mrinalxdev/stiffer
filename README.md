# Stiffer 

This layer is a go nats jetstream work queue DAG. Each stage is a durable consumer on its own subject with its own concurrency so we scale 'analyze' independently of ingest / proxy and index
