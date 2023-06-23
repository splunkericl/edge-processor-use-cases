## Description

Currently, [Edge Processor](https://docs.splunk.com/Documentation/SplunkCloud/9.0.2303/EdgeProcessor/AboutEdgeProcessorSolution)(EP) doesn't have native capabilities to send S3 data to EP. This sub repository provides aws lambda code any customers can use to send their S3 data to EP.

### Architecture

[S3] ----(S3 Trigger)-----> [AWS Lambda(fetch S3 content by bucket and key)] ------(HTTP call)----> EP

## How to Use

### Pre-req

- You have already set up a EP and deployed it on your host successfully
- Your EP is running and healthy and HEC receiver is enabled
- You already have S3 bucket with contents inside that you want to send to EP
- You VPC/Security group allow HTTP traffic between AWS Lambda and your EP on your HEC port

### Steps

1. Clone this repo and run the script to build and zip the code: `bash ./buildZip.sh`. This will create a zip file `s3-to-ep.zip`.
2. Follow the [guide](https://docs.aws.amazon.com/lambda/latest/dg/with-s3-example.html#with-s3-example-create-policy) to create a policy to allow fetching from S3 and adding cloud watch logs.
3. Follow the [guide](https://docs.aws.amazon.com/lambda/latest/dg/with-s3-example.html#with-s3-example-create-role) to create an execution role to use the policy from step 2.
4. You can follow the [guide](https://docs.aws.amazon.com/lambda/latest/dg/with-s3-example.html#with-s3-example-create-function) to create a Lambda function. Note:
   - You can use your own **Function name**
   - **Runtime** is set to ``Go 1.x``
   - **Architect** is set to ``x86_64``
5. On the **code** tab, go to **Code Source** section. Press **Upload From** and attach the zip file from step 1.
6. After uploading the Lambda function, go to **Runtime settings**. Verify **Handler** is set to ``main``. If not, edit it. 
7. You can follow the [guide](https://docs.aws.amazon.com/lambda/latest/dg/with-s3-example.html#with-s3-example-create-trigger) to set up the S3 trigger. Note:
   - For moving archived data, you can configure to only use `COPY` **Event types**
   - You can configure other event types like `All object create events` or `POST` but the tool isn't fullly tested to stream data continuously
8. Under **Configuration** tab, go to **Environment variables**. Follow the [environment variable sections](#environment-variables) and add ones for your systems.
9. Follow the [guide](https://docs.aws.amazon.com/lambda/latest/dg/with-s3-example.html#with-s3-example-test-dummy-event) to test the function with a dummy event. Note:
10. Check the log to see if it was successful. If it is, check the dashboard to see if EP has received the event.
11. Publish the Lambda function

#### Use Case - Route Archived Data to EP

After following the [steps](#steps) to set up your Lambda function:
1. Go to your S3 Bucket page. Select the folder containing the contents you wish to route into EP
2. Click **Action** and select **Copy** in the dropdown
3. Click **Browse S3** and select the bucket you set up S3 trigger with.
4. Click **Copy**
5. Your data will be copied into the bucket and subsequently routed to Lambda and then EP.

### Environment Variables

| Key                 | Description                                                                                                                                                                                 | Required | Example Value                                                 |
|---------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------|---------------------------------------------------------------|
| EDGE_PROCESSOR_HOST | The EP host that AWS Lambda will connect to including the port.                                                                                                                             | Yes      | http://ec2-26-78-145-255.us-west-2.compute.amazonaws.com:8088 |
| TLS_CLIENT_CERT     | The client certificate to use if connection is TLS. If not provided along with `TLS_CLIENT_KEY`, it won't enable TLS.                                                                       | No       |                                                               |
| TLS_CLIENT_KEY      | The client private key to use if connection is TLS. If not provided along with `TLS_CLIENT_CERT`, it won't enable TLS.                                                                      | No       |                                                               |
| TLS_CLIENT_CA_CERT  | The custom CA cert to use for TLS. It will be appended on top of system certs.                                                                                                              | No       |                                                               |
| ENCODING_METHOD     | Set this field if s3 content is encoded in provided method. Note: currently only support gzip.                                                                                              | No       | gzip                                                          |
| EVENT_SOURCETYPE    | If set, event sent to EP will use provided sourcetype. if not set, defaults to `archived_data`                                                                                              | No       | test-sourcetype                                               |
| EVENT_INDEX         | If set, event sent to EP will use provided index. if not set, defaults to `main`                                                                                                            | No       | event-index                                                   |
| EVENT_IS_RAW        | If set, event will be sent to EP raw endpoint. Note: this is unofficial support and line breaking is made best efforts. User configured line brekaing in EP won't apply. default to `false` | No       | true                                                          |

### Limitation

Here are some limitations:
- S3 Content
  - if content is encoded, only GZIP encoded format is supported for now
  - if content is in parquet format, it won't be parsed properly
- Build/Zip tool isn't tested on windows
- EP can't use users provided line breaking configurations for HEC raw data.

## Development

Issues are welcome to be created for feature. Pull requests are welcomed for improvement.

### Testing

Run unit test by executing `go test ./...`