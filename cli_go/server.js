const express = require('express');
const http = require('http');
const WebSocket = require('ws');
const pty = require('node-pty');
const os = require('os');

const app = express();
const server = http.createServer(app);
const wss = new WebSocket.Server({ server });

// Serve static files
app.use(express.static('.'));

// Serve the HTML page
app.get('/', (req, res) => {
  res.sendFile(__dirname + '/index.html');
});

// WebSocket connection handler
wss.on('connection', (ws) => {
  console.log('Client connected');

  // Determine shell based on OS
  const shell = os.platform() === 'win32' ? 'powershell.exe' : 'bash';
  
  // Spawn a shell process at home directory
  const ptyProcess = pty.spawn(shell, [], {
    name: 'xterm-color',
    cols: 80,
    rows: 30,
    cwd: os.homedir(),
    env: process.env
  });

  // Send shell output to the client
  ptyProcess.onData((data) => {
    ws.send(data);
  });

  // Handle resize events
  ws.on('message', (msg) => {
    try {
      const data = JSON.parse(msg);
      if (data.type === 'resize') {
        ptyProcess.resize(data.cols, data.rows);
      } else {
        ptyProcess.write(msg);
      }
    } catch (e) {
      // If not JSON, treat as regular input
      ptyProcess.write(msg);
    }
  });

  // Clean up on disconnect
  ws.on('close', () => {
    console.log('Client disconnected');
    ptyProcess.kill();
  });

  // Handle shell exit
  ptyProcess.onExit(() => {
    ws.close();
  });
});

const PORT = 3000;
server.listen(PORT, () => {
  console.log(`Terminal server running on http://localhost:${PORT}`);
});
