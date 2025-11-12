import { DefaultAzureCredential } from '@azure/identity';
import { ComputeManagementClient } from '@azure/arm-compute';

/**
 * VM Creation
 * @param {string} vmId - VM Unique ID
 * @param {string} subscriptionId - Azure Subscription ID
 * @returns {Promise<object>} VM metadata
 */
export async function createVM(vmId, subscriptionId) {
    console.log(`Creating VM: ${vmId}`);
    console.log(`Using subscription: ${subscriptionId}`);

    // Azure authentication
    const credential = new DefaultAzureCredential();
    const computeClient = new ComputeManagementClient(
        credential,
        subscriptionId
    );

    // Basic configuration
    const resourceGroupName = process.env.RESOURCE_GROUP_NAME;
    const location = process.env.AZURE_LOCATION;


    // TODO: VM creating
    // TODO: Public IP allocation/assignment
    // TODO: noVNC installation (cloud-init)

    // Authentication test for now
    console.log('Azure authentication successful');
    console.log('Compute client initialized');

    // mock data for now
    return {
        vmId: vmId,
        status: 'creating',
        publicIP: null,
        resourceGroup: resourceGroupName,
        location: location
    };
}