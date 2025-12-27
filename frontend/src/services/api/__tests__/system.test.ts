import { describe, it, expect, vi, beforeEach } from 'vitest';
import * as system from '../system';

// Mock the apiClient
vi.mock('../../apiClient', () => ({
  get: vi.fn(),
  post: vi.fn(),
}));

import { get, post } from '../../apiClient';

describe('system API', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('getStatus', () => {
    it('should call get with correct path', async () => {
      const mockStatus = { connected: true };
      vi.mocked(get).mockResolvedValue(mockStatus);

      const result = await system.getStatus();

      expect(get).toHaveBeenCalledWith('/system/status');
      expect(result).toEqual(mockStatus);
    });
  });

  describe('getDaemonStatus', () => {
    it('should call get with correct path', async () => {
      const mockStatus = { status: 'running', connected: true };
      vi.mocked(get).mockResolvedValue(mockStatus);

      const result = await system.getDaemonStatus();

      expect(get).toHaveBeenCalledWith('/system/daemon/status');
      expect(result).toEqual(mockStatus);
    });
  });

  describe('connectDaemon', () => {
    it('should call post with correct path', async () => {
      const mockResult = { status: 'connected' };
      vi.mocked(post).mockResolvedValue(mockResult);

      const result = await system.connectDaemon();

      expect(post).toHaveBeenCalledWith('/system/daemon/connect');
      expect(result).toEqual(mockResult);
    });
  });

  describe('disconnectDaemon', () => {
    it('should call post with correct path', async () => {
      const mockResult = { status: 'disconnected' };
      vi.mocked(post).mockResolvedValue(mockResult);

      const result = await system.disconnectDaemon();

      expect(post).toHaveBeenCalledWith('/system/daemon/disconnect');
      expect(result).toEqual(mockResult);
    });
  });

  describe('getVersion', () => {
    it('should call get with correct path', async () => {
      const mockVersion = { version: '1.0.0', service: 'mtga-companion' };
      vi.mocked(get).mockResolvedValue(mockVersion);

      const result = await system.getVersion();

      expect(get).toHaveBeenCalledWith('/system/version');
      expect(result).toEqual(mockVersion);
    });
  });

  describe('getDatabasePath', () => {
    it('should call get with correct path', async () => {
      const mockPath = { path: '/path/to/db' };
      vi.mocked(get).mockResolvedValue(mockPath);

      const result = await system.getDatabasePath();

      expect(get).toHaveBeenCalledWith('/system/database/path');
      expect(result).toEqual(mockPath);
    });
  });

  describe('setDatabasePath', () => {
    it('should call post with path', async () => {
      const mockResult = { status: 'ok' };
      vi.mocked(post).mockResolvedValue(mockResult);

      const result = await system.setDatabasePath('/new/path');

      expect(post).toHaveBeenCalledWith('/system/database/path', { path: '/new/path' });
      expect(result).toEqual(mockResult);
    });
  });

  describe('getCurrentAccount', () => {
    it('should call get with correct path', async () => {
      const mockAccount = { id: 123, name: 'Player' };
      vi.mocked(get).mockResolvedValue(mockAccount);

      const result = await system.getCurrentAccount();

      expect(get).toHaveBeenCalledWith('/system/account');
      expect(result).toEqual(mockAccount);
    });
  });

  describe('clearAllData', () => {
    it('should call post with correct path', async () => {
      vi.mocked(post).mockResolvedValue(undefined);

      await system.clearAllData();

      expect(post).toHaveBeenCalledWith('/export/clear');
    });
  });

  describe('checkOllamaStatus', () => {
    it('should call post with endpoint and model', async () => {
      const mockStatus = { available: true, modelReady: true };
      vi.mocked(post).mockResolvedValue(mockStatus);

      const result = await system.checkOllamaStatus('http://localhost:11434', 'llama2');

      expect(post).toHaveBeenCalledWith('/llm/status', {
        endpoint: 'http://localhost:11434',
        model: 'llama2',
      });
      expect(result).toEqual(mockStatus);
    });
  });

  describe('getAvailableOllamaModels', () => {
    it('should call get with endpoint in query', async () => {
      const mockModels = [{ name: 'llama2', size: 1000 }];
      vi.mocked(get).mockResolvedValue(mockModels);

      const result = await system.getAvailableOllamaModels('http://localhost:11434');

      expect(get).toHaveBeenCalledWith('/llm/models?endpoint=http%3A%2F%2Flocalhost%3A11434');
      expect(result).toEqual(mockModels);
    });

    it('should call get without query when endpoint is empty', async () => {
      const mockModels = [{ name: 'llama2', size: 1000 }];
      vi.mocked(get).mockResolvedValue(mockModels);

      const result = await system.getAvailableOllamaModels('');

      expect(get).toHaveBeenCalledWith('/llm/models');
      expect(result).toEqual(mockModels);
    });
  });

  describe('pullOllamaModel', () => {
    it('should call post with endpoint and model', async () => {
      vi.mocked(post).mockResolvedValue(undefined);

      await system.pullOllamaModel('http://localhost:11434', 'llama2');

      expect(post).toHaveBeenCalledWith('/llm/models/pull', {
        endpoint: 'http://localhost:11434',
        model: 'llama2',
      });
    });
  });

  describe('testLLMGeneration', () => {
    it('should call post and return response', async () => {
      vi.mocked(post).mockResolvedValue({ response: 'Hello from LLM!' });

      const result = await system.testLLMGeneration('http://localhost:11434', 'llama2');

      expect(post).toHaveBeenCalledWith('/llm/test', {
        endpoint: 'http://localhost:11434',
        model: 'llama2',
      });
      expect(result).toBe('Hello from LLM!');
    });
  });

  describe('exportMLTrainingData', () => {
    it('should call get with limit in query', async () => {
      const mockData = { records: [] };
      vi.mocked(get).mockResolvedValue(mockData);

      const result = await system.exportMLTrainingData(100);

      expect(get).toHaveBeenCalledWith('/feedback/ml-training?limit=100');
      expect(result).toEqual(mockData);
    });
  });
});
