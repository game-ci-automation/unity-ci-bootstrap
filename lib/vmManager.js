import { DefaultAzureCredential } from '@azure/identity';
import { ComputeManagementClient } from '@azure/arm-compute';
import { NetworkManagementClient } from '@azure/arm-network';
import { ResourceManagementClient } from '@azure/arm-resources';

/**
 * Ensure Resource Group exists, create if not
 * @param {ResourceManagementClient} resourceClient - Azure Resource Management client
 * @param {string} resourceGroupName - Name of the resource group
 * @param {string} location - Azure region (e.g., 'eastus')
 * @returns {Promise<void>} - No return value, throws on error
 */
async function ensureResourceGroup(resourceClient, resourceGroupName, location) {
    console.log(`Checking resource group: ${resourceGroupName}`);

    try {
        await resourceClient.resourceGroups.get(resourceGroupName);
        console.log(`Resource group ${resourceGroupName} already exists`);
    } catch (error) {
        if (error.statusCode === 404) {
            console.log(`Creating resource group ${resourceGroupName} in ${location}...`);
            await resourceClient.resourceGroups.createOrUpdate(resourceGroupName, {
                location: location
            });
            console.log(`Resource group ${resourceGroupName} created`);
        } else {
            throw error;
        }
    }
}

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
    const resourceClient = new ResourceManagementClient(
        credential,
        subscriptionId
    );

    // Basic configuration
    const resourceGroupName = process.env.RESOURCE_GROUP_NAME;
    const location = process.env.AZURE_LOCATION;


    // TODO: VM creating
    try {
        await ensureResourceGroup(resourceClient, resourceGroupName, location);

        // TODO: Public IP allocation/assignment
        // TODO: noVNC installation (cloud-init)

        return {
            vmId: vmId,
            status: 'resource-group-ready',
            resourceGroup: resourceGroupName,
            location: location
        };

    } catch(error) {
        console.error('VM creation error', error);
        throw error;
    }
}