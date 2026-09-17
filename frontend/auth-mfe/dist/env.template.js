// Environment configuration — values substituted at container startup
// ${API_BASE}, ${IS_STAGING}, ${ENV_NAME} are replaced by envsubst
(function() {
    var apiBase = '${API_BASE}';
    var envName = '${ENV_NAME}';
    
    window.__PAWRADISE_ENV__ = {
        apiBase: apiBase || '/api/v1',
        envName: envName
    };
})();
