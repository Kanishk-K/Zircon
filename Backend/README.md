# Zircon Backend

Zircon's backend is written in Go and requires Docker to run some of the backend operations such as job queues or video generation. In addition, do make sure to have an AWS account setup on your system so that the codebase can access AWS resources.

## Dependencies
- Golang >= 1.23.4
- [AWS SAM](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html)
- Docker
