/**
 * Base API client for REST API communication.
 * Replaces direct Wails function calls with HTTP requests.
 */

export interface ApiConfig {
  baseUrl: string;
  timeout?: number;
}

export interface ApiResponse<T> {
  data: T;
  error?: string;
}

export interface ApiError {
  error: string;
  code?: string;
  details?: string;
}

// Default configuration - can be overridden
let config: ApiConfig = {
  baseUrl: 'http://localhost:8080/api/v1',
  timeout: 30000,
};

/**
 * Configure the API client.
 */
export function configureApi(newConfig: Partial<ApiConfig>): void {
  config = { ...config, ...newConfig };
}

/**
 * Get the current API configuration.
 */
export function getApiConfig(): ApiConfig {
  return { ...config };
}

/**
 * Custom error class for API errors.
 */
export class ApiRequestError extends Error {
  public readonly status: number;
  public readonly code?: string;
  public readonly details?: string;

  constructor(message: string, status: number, code?: string, details?: string) {
    super(message);
    this.name = 'ApiRequestError';
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

/**
 * Make an HTTP request to the API.
 */
async function request<T>(
  method: string,
  path: string,
  body?: unknown,
  options: RequestInit = {}
): Promise<T> {
  const url = `${config.baseUrl}${path}`;

  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), config.timeout);

  try {
    const response = await fetch(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      body: body ? JSON.stringify(body) : undefined,
      signal: controller.signal,
      ...options,
    });

    clearTimeout(timeoutId);

    if (!response.ok) {
      let errorData: ApiError = { error: 'Unknown error' };
      try {
        errorData = await response.json();
      } catch {
        errorData = { error: response.statusText || 'Request failed' };
      }
      throw new ApiRequestError(
        errorData.error,
        response.status,
        errorData.code,
        errorData.details
      );
    }

    // Handle 204 No Content
    if (response.status === 204) {
      return undefined as T;
    }

    const data = await response.json();
    return data.data as T;
  } catch (error) {
    clearTimeout(timeoutId);

    if (error instanceof ApiRequestError) {
      throw error;
    }

    if (error instanceof Error) {
      if (error.name === 'AbortError') {
        throw new ApiRequestError('Request timeout', 408);
      }
      throw new ApiRequestError(error.message, 0);
    }

    throw new ApiRequestError('Unknown error', 0);
  }
}

/**
 * HTTP GET request.
 */
export function get<T>(path: string, options?: RequestInit): Promise<T> {
  return request<T>('GET', path, undefined, options);
}

/**
 * HTTP POST request.
 */
export function post<T>(path: string, body?: unknown, options?: RequestInit): Promise<T> {
  return request<T>('POST', path, body, options);
}

/**
 * HTTP PUT request.
 */
export function put<T>(path: string, body?: unknown, options?: RequestInit): Promise<T> {
  return request<T>('PUT', path, body, options);
}

/**
 * HTTP PATCH request.
 */
export function patch<T>(path: string, body?: unknown, options?: RequestInit): Promise<T> {
  return request<T>('PATCH', path, body, options);
}

/**
 * HTTP DELETE request.
 */
export function del<T>(path: string, options?: RequestInit): Promise<T> {
  return request<T>('DELETE', path, undefined, options);
}

/**
 * Check if the API server is reachable.
 */
export async function healthCheck(): Promise<boolean> {
  try {
    const response = await fetch(`${config.baseUrl.replace('/api/v1', '')}/health`, {
      method: 'GET',
    });
    return response.ok;
  } catch {
    return false;
  }
}
