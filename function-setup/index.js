module.exports = async function (context, req) {
    context.log('Setup function called');

    const html = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Unity CI License Activator</title>
    <style>
        body {
            font-family: system-ui, -apple-system, sans-serif;
            max-width: 600px;
            margin: 50px auto;
            padding: 20px;
            line-height: 1.6;
        }
        .status {
            padding: 20px;
            background: #f0f0f0;
            border-radius: 8px;
            margin: 20px 0;
        }
        .loading {
            color: #666;
        }
    </style>
</head>
<body>
    <h1>Unity CI License Activator</h1>

    <div class="status">
        <h2>Status: VM Creating</h2>
        <p class="loading">⏳ Provisioning virtual machine...</p>
        <p>This may take 5-8 minutes. Please wait.</p>
    </div>

    <div id="details">
        <p><strong>What's happening:</strong></p>
        <ul>
            <li>Creating Ubuntu Desktop VM</li>
            <li>Installing noVNC for web access</li>
            <li>Downloading Unity Hub installer</li>
        </ul>
    </div>
</body>
</html>
    `;

    context.res = {
        status: 200,
        headers: {
            'Content-Type': 'text/html; charset=utf-8'
        },
        body: html
    };
};