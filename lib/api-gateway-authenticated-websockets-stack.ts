import * as cdk from 'aws-cdk-lib';
import {CfnOutput, CfnParameter} from 'aws-cdk-lib';
import {Construct} from 'constructs';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import {AttributeType} from 'aws-cdk-lib/aws-dynamodb';
import * as cognito from 'aws-cdk-lib/aws-cognito';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as api_gw_v2 from 'aws-cdk-lib/aws-apigatewayv2';
import {WebSocketLambdaIntegration} from 'aws-cdk-lib/aws-apigatewayv2-integrations';

export class ApiGatewayAuthenticatedWebsocketsStack extends cdk.Stack {

    readonly stage: api_gw_v2.WebSocketStage;

    constructor(scope: Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        const userPoolId = new CfnParameter(this, 'userPoolId', {
            description: 'ID of the Cognito User Pool',
            type: 'String'
        }).valueAsString

        const connectionsTable = new dynamodb.Table(this, 'Connections', {
            tableName: 'connections',
            partitionKey: {
                type: AttributeType.STRING,
                name: 'connection_id'
            },
        })
        connectionsTable.addGlobalSecondaryIndex({
            indexName: 'task_id-index',
            partitionKey: {
                type: AttributeType.STRING,
                name: 'task_id'
            }
        })

        const connectLambda = new lambda.Function(this, 'Connect', {
            functionName: 'WebSocketConnect',
            handler: 'bootstrap',
            runtime: lambda.Runtime.PROVIDED_AL2023,
            code: lambda.Code.fromAsset('./lib/connect/connect.zip'),
        });
        connectionsTable.grantWriteData(connectLambda);

        const disconnectLambda = new lambda.Function(this, 'Disconnect', {
            functionName: 'WebSocketDisconnect',
            handler: 'bootstrap',
            runtime: lambda.Runtime.PROVIDED_AL2023,
            code: lambda.Code.fromAsset('./lib/disconnect/disconnect.zip'),
        });
        connectionsTable.grantWriteData(disconnectLambda);

        const userPool = cognito.UserPool.fromUserPoolId(this, 'UserPool', userPoolId);
        const authorizeLambda = new lambda.Function(this, 'Authorize', {
            functionName: 'WebSocketAuthorize',
            handler: 'bootstrap',
            runtime: lambda.Runtime.PROVIDED_AL2023,
            code: lambda.Code.fromAsset('./lib/authorize/authorize.zip'),
            environment: {
                'JWKS_URL': `https://cognito-idp.${userPool.env.region}.amazonaws.com/${userPoolId}/.well-known/jwks.json`
            }
        });
        connectionsTable.grantWriteData(authorizeLambda);

        const apiGateway = new api_gw_v2.WebSocketApi(this, 'WebSocketAPI', {
            apiName: 'tasks',
            connectRouteOptions: {
                integration: new WebSocketLambdaIntegration('Connect', connectLambda)
            },
            disconnectRouteOptions: {
                integration: new WebSocketLambdaIntegration('Disconnect', disconnectLambda)
            },
            defaultRouteOptions: {
                integration: new WebSocketLambdaIntegration('Authorize', authorizeLambda)
            }
        });
        this.stage = new api_gw_v2.WebSocketStage(this, 'Stage', {
            stageName: 'development',
            webSocketApi: apiGateway,
            autoDeploy: true
        });

        const sendLambda = new lambda.Function(this, 'Send', {
            functionName: 'WebSocketSend',
            handler: 'bootstrap',
            runtime: lambda.Runtime.PROVIDED_AL2023,
            code: lambda.Code.fromAsset('./lib/send/send.zip'),
            environment: {
                'CONNECTIONS_URL': this.stage.callbackUrl
            }
        });
        connectionsTable.grantReadData(sendLambda);
        apiGateway.grantManageConnections(sendLambda);

        new CfnOutput(this, 'WebSocketURL', {
            key: 'WebSocketURL',
            value: this.stage.url
        });
    }
}
