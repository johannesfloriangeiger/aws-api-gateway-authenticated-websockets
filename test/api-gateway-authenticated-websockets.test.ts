import * as cdk from 'aws-cdk-lib';
import {Template} from 'aws-cdk-lib/assertions';
import * as ApiGatewayAuthenticatedWebsockets from '../lib/api-gateway-authenticated-websockets-stack';

test('Stack created', () => {
    const app = new cdk.App();
    const stack = new ApiGatewayAuthenticatedWebsockets.ApiGatewayAuthenticatedWebsocketsStack(app, 'MyTestStack');
    const template = Template.fromStack(stack);

    template.hasResourceProperties('AWS::DynamoDB::Table', {
        AttributeDefinitions: [{
            AttributeName: 'connection_id',
            AttributeType: 'S',
        }, {
            AttributeName: 'task_id',
            AttributeType: 'S',
        }],
        GlobalSecondaryIndexes: [{
            IndexName: 'task_id-index',
            KeySchema: [{
                AttributeName: 'task_id',
            }],
        }],
        TableName: 'connections'
    });

    template.hasResourceProperties('AWS::Lambda::Function', {
        FunctionName: 'WebSocketConnect',
        Handler: 'bootstrap',
        Runtime: 'provided.al2023'
    });

    template.hasResourceProperties('AWS::Lambda::Function', {
        FunctionName: 'WebSocketDisconnect',
        Handler: 'bootstrap',
        Runtime: 'provided.al2023'
    });

    template.hasResourceProperties('AWS::Lambda::Function', {
        FunctionName: 'WebSocketSend',
        Handler: 'bootstrap',
        Runtime: 'provided.al2023',
        Environment: {
            'Variables': {
                'CONNECTIONS_URL': {
                    'Fn::Join': ['', ['https://', stack.resolve(stack.stage.api.apiId), '.execute-api.', stack.resolve(stack.stage.api.env.region), '.', {'Ref': 'AWS::URLSuffix'}, '/development']]
                }
            }
        }
    });
});
