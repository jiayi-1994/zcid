import '@testing-library/jest-dom';
import { vi } from 'vitest';

Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

const storage = new Map<string, string>();
Object.defineProperty(window, 'localStorage', {
  configurable: true,
  value: {
    getItem: vi.fn((key: string) => storage.get(key) ?? null),
    setItem: vi.fn((key: string, value: string) => { storage.set(key, value); }),
    removeItem: vi.fn((key: string) => { storage.delete(key); }),
    clear: vi.fn(() => { storage.clear(); }),
    key: vi.fn((index: number) => Array.from(storage.keys())[index] ?? null),
    get length() { return storage.size; },
  },
});

// Mock arco-design Message to avoid ReactDOM.render issue in React 19
vi.mock('@arco-design/web-react', async () => {
  const actual = await vi.importActual('@arco-design/web-react');
  return {
    ...actual,
    Message: {
      success: vi.fn(),
      error: vi.fn(),
      warning: vi.fn(),
      info: vi.fn(),
      loading: vi.fn(),
      clear: vi.fn(),
    },
  };
});
