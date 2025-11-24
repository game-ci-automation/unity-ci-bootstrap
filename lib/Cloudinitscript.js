import { readFileSync } from 'fs';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';
import { config } from './config.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

/**
 * Generate cloud-init script for Ubuntu Desktop + VNC + noVNC setup
 * @param {string} vncPassword - VNC connection password
 * @returns {string} Base64 encoded cloud-init script
 */
export function generateCloudInitScript(
    vncPassword = config.vnc.password,
    username = config.vm.adminUsername) {
    // Load YAML template
    const templatePath = join(__dirname, 'cloud-init-template.yaml');
    let template = readFileSync(templatePath, 'utf8');

    // Replace password placeholder
    template = template.replace(/{{VNC_PASSWORD}}/g, vncPassword);
    template = template.replace(/{{USERNAME}}/g, username);

    // Return Base64 encoded script
    return Buffer.from(template).toString('base64');
}