import { DefaultAzureCredential } from '@azure/identity';
import { ComputeManagementClient } from '@azure/arm-compute';

/**
 * VM Creation
 * @param {string} vmId - VM Unique ID
 * @returns {Promise<object>} VM metadata
 */
export async function createVM(vmId) {
    console.log(`Creating VM: ${vmId}`);

    // TODO: Azure authentication
    // TODO: VM creating
    // TODO: Public IP allocation/assignment
    // TODO: noVNC installation (cloud-init)

    // mock data for now
    return {
        vmId: vmId,
        status: 'creating',
        publicIP: null
    };
}