import { ReactElement } from 'react';
import { render, RenderOptions } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';

interface CustomRenderOptions extends Omit<RenderOptions, 'wrapper'> {
  initialRoute?: string;
}

function AllTheProviders({ children }: { children: React.ReactNode }) {
  return <BrowserRouter>{children}</BrowserRouter>;
}

export function renderWithRouter(
  ui: ReactElement,
  options?: CustomRenderOptions
) {
  const { initialRoute = '/', ...renderOptions } = options || {};

  if (initialRoute !== '/') {
    window.history.pushState({}, 'Test page', initialRoute);
  }

  return render(ui, { wrapper: AllTheProviders, ...renderOptions });
}

export * from '@testing-library/react';
export { renderWithRouter as render };
