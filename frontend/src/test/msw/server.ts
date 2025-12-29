/**
 * MSW server setup for integration tests.
 */
import { setupServer } from 'msw/node';
import { handlers } from './handlers';

// Create the MSW server with default handlers
export const server = setupServer(...handlers);
