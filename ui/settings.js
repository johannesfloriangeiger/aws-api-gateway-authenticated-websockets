window.config = {
    BASE_URL: `${window.location.protocol}//${window.location.host}`,
    COGNITO: {
        RESPONSE_TYPE: {
            TOKEN: 'token'
        },
        GRANT_TYPE: {
            AUTHORIZATION_CODE: 'authorization_code',
        },
        URL: '${COGNITO_URL}',
        CLIENT: {
            ID: '${COGNITO_CLIENT_ID}',
        },
        ID_TOKEN: 'id_token'
    },
    WEB_SOCKET: {
        URL: '${WEB_SOCKET_URL}?taskId=1'
    }
}