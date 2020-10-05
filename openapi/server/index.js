const httpProxy = require('http-proxy');
const http = require('http');
const url = require('url');
const finalhandler = require('finalhandler')
const serveStatic = require('serve-static')

// Proxy targets
const targets = {
    '/community': 'http://localhost:8081',
    '/cloud': 'http://localhost:8082'
};

const proxy = httpProxy.createServer();

// Generic error handler
proxy.on('error', function(err, req, res) {
    res.writeHead(500);
    res.end('Internal Server Error on Proxy\n');
});


const staticServer = serveStatic('./www', { 'index': ['index.html'] })

http.createServer((req, res) => {
    const path = url.parse(req.url).pathname;

    for (const target in targets) {
        if (path.startsWith(target)) {
            const targetUrl = `${targets[target]}/${path.replace(target, '')}` // rewrite URL removing target prefix
            return proxy.web(req, res, { target: targetUrl, ignorePath: true });
        }
    }

    staticServer(req, res, finalhandler(req, res))
}).listen(8080);

console.log('Proxy running on port 8080');
