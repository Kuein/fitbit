# fitbit RPG

This is serverless application for FitBit RPG 

## Requirements

* AWS CLI already configured with Administrator permission
* [Golang](https://golang.org)
* SAM CLI - [Install the SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html)
* DynamoDB table with name "Fitbit", id key == "primary" and sort key == "secondary"
* Fitbit application with Server-type of OAuth

## Setup process

### Installing dependencies & building the target 

Read more about [SAM Build here](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/sam-cli-command-reference-sam-build.html) 

The `sam build` command is wrapped inside of the `Makefile`. 

To deploy your application for the first time, run the following in your shell:

```bash
sam deploy --guided
```

The command will package and deploy your application to AWS, with a series of prompts:

* **Stack Name**: The name of the stack to deploy to CloudFormation. This should be unique to your account and region, and a good starting point would be something matching your project name.
* **AWS Region**: The AWS region you want to deploy your app to.
* **Confirm changes before deploy**: If set to yes, any change sets will be shown to you before execution for manual review. If set to no, the AWS SAM CLI will automatically deploy application changes.
* **Allow SAM CLI IAM role creation**: Many AWS SAM templates, including this example, create AWS IAM roles required for the AWS Lambda function(s) included to access AWS services. By default, these are scoped down to minimum required permissions. To deploy an AWS CloudFormation stack which creates or modifies IAM roles, the `CAPABILITY_IAM` value for `capabilities` must be provided. If permission isn't provided through this prompt, to deploy this example you must explicitly pass `--capabilities CAPABILITY_IAM` to the `sam deploy` command.
* **Save arguments to samconfig.toml**: If set to yes, your choices will be saved to a configuration file inside the project, so that in the future you can just re-run `sam deploy` without parameters to deploy changes to your application.

You can find your API Gateway Endpoint URL in the output values displayed after deployment.
Future deployment will be available just as 
```bash
make deploy`
```
# Appendix

### Golang installation

Please ensure Go 1.x (where 'x' is the latest version) is installed as per the instructions on the official golang website: https://golang.org/doc/install

A quickstart way would be to use Homebrew, chocolatey or your linux package manager.

## Bringing to the next level current implementation

- Full scale authentication. Right now its primitive cookie (without expiration) and passwords saved in database as plain text
- Front-end on javascript framework or WebAssembly. Currently we are using server-side template rendering
- Optimization for database access and template rendering with `sync.Once`
- Optimization access to fitbit API for refreshing tokens only if this is really needed
 
## Bringing to the next level project

- Actual game mechanics something like https://www.youtube.com/watch?v=olguFFLA_7g (MyBrute)
- Complex user profiles with customization option for character
- 
