import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  configureApi,
  getApiConfig,
  get,
  post,
  put,
  del,
  healthCheck,
  ApiRequestError,
} from '../apiClient';

// Mock fetch globally
const mockFetch = vi.fn();
global.fetch = mockFetch;

describe('apiClient', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Reset to default config
    configureApi({
      baseUrl: 'http://localhost:8080/api/v1',
      timeout: 30000,
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('configureApi', () => {
    it('should update API configuration', () => {
      configureApi({ baseUrl: 'http://example.com/api' });
      const config = getApiConfig();
      expect(config.baseUrl).toBe('http://example.com/api');
    });

    it('should merge with existing config', () => {
      configureApi({ timeout: 5000 });
      const config = getApiConfig();
      expect(config.timeout).toBe(5000);
      expect(config.baseUrl).toBe('http://localhost:8080/api/v1');
    });
  });

  describe('get', () => {
    it('should make a GET request', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ data: { id: 1, name: 'Test' } }),
      });

      const result = await get('/test');

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/test',
        expect.objectContaining({
          method: 'GET',
          headers: expect.objectContaining({
            'Content-Type': 'application/json',
          }),
        })
      );
      expect(result).toEqual({ id: 1, name: 'Test' });
    });

    it('should throw ApiRequestError on 404', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 404,
        statusText: 'Not Found',
        json: async () => ({ error: 'Resource not found' }),
      });

      try {
        await get('/nonexistent');
        expect.fail('Should have thrown');
      } catch (error) {
        expect(error).toBeInstanceOf(ApiRequestError);
        expect((error as ApiRequestError).status).toBe(404);
        expect((error as ApiRequestError).message).toBe('Resource not found');
      }
    });

    it('should throw ApiRequestError on 500', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        json: async () => ({ error: 'Server error', code: 'INTERNAL' }),
      });

      await expect(get('/error')).rejects.toThrow(ApiRequestError);
    });
  });

  describe('post', () => {
    it('should make a POST request with body', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ data: { success: true } }),
      });

      const body = { name: 'Test', value: 123 };
      const result = await post('/create', body);

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/create',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(body),
        })
      );
      expect(result).toEqual({ success: true });
    });

    it('should handle 201 Created response', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 201,
        json: async () => ({ data: { id: 'new-id' } }),
      });

      const result = await post('/items', { name: 'New Item' });
      expect(result).toEqual({ id: 'new-id' });
    });
  });

  describe('put', () => {
    it('should make a PUT request', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ data: { updated: true } }),
      });

      const result = await put('/items/1', { name: 'Updated' });

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/items/1',
        expect.objectContaining({ method: 'PUT' })
      );
      expect(result).toEqual({ updated: true });
    });
  });

  describe('del', () => {
    it('should make a DELETE request', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 204,
      });

      const result = await del('/items/1');

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/items/1',
        expect.objectContaining({ method: 'DELETE' })
      );
      expect(result).toBeUndefined();
    });
  });

  describe('healthCheck', () => {
    it('should return true when server is healthy', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
      });

      const result = await healthCheck();
      expect(result).toBe(true);
    });

    it('should return false when server is unreachable', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'));

      const result = await healthCheck();
      expect(result).toBe(false);
    });

    it('should return false when server returns error', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 503,
      });

      const result = await healthCheck();
      expect(result).toBe(false);
    });
  });

  describe('error handling', () => {
    it('should handle network errors', async () => {
      mockFetch.mockRejectedValue(new Error('Network error'));

      try {
        await get('/test');
        expect.fail('Should have thrown');
      } catch (error) {
        expect(error).toBeInstanceOf(ApiRequestError);
        expect((error as ApiRequestError).status).toBe(0);
        expect((error as ApiRequestError).message).toBe('Network error');
      }
    });

    it('should handle JSON parse errors in error response', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        json: async () => {
          throw new Error('Invalid JSON');
        },
      });

      try {
        await get('/test');
        expect.fail('Should have thrown');
      } catch (error) {
        expect(error).toBeInstanceOf(ApiRequestError);
        expect((error as ApiRequestError).status).toBe(500);
        expect((error as ApiRequestError).message).toBe('Internal Server Error');
      }
    });
  });
});

describe('ApiRequestError', () => {
  it('should create error with all properties', () => {
    const error = new ApiRequestError('Not found', 404, 'NOT_FOUND', 'Resource does not exist');

    expect(error.message).toBe('Not found');
    expect(error.status).toBe(404);
    expect(error.code).toBe('NOT_FOUND');
    expect(error.details).toBe('Resource does not exist');
    expect(error.name).toBe('ApiRequestError');
  });

  it('should be instanceof Error', () => {
    const error = new ApiRequestError('Test', 500);
    expect(error).toBeInstanceOf(Error);
    expect(error).toBeInstanceOf(ApiRequestError);
  });
});
