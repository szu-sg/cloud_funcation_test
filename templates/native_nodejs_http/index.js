const http = require('http');

const server = http.createServer((req, res) => {
  res.statusCode = 200;
  res.setHeader('Content-Type', 'text/plain');
  res.end('Hello, World!');
});

server.listen(process.env.PORT || 3000, 'localhost', () => {
  console.log(`server listen at: ${process.env.PORT || 3000}`);
});
