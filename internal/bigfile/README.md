# bigfile

## Scenario

- Show how Temporal can be use to orchestrate; where normal ingress API gateway + nginx 

- It can safely break up content to be uploaded or downloaded in parallel.
Utilise S3 atomic check if exist primitive; ability to select chunks and also its signed url for this.

- Show it can work for other S3-like env like minio + R2

- Show how unit test can be used to simulate all the weird edge cases that might be faced and it can continue in the face of disruption

