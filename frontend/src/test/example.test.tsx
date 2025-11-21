import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';

describe('Example Test', () => {
  it('should render a simple component', () => {
    const TestComponent = () => <div>Hello Test</div>;
    render(<TestComponent />);
    expect(screen.getByText('Hello Test')).toBeInTheDocument();
  });

  it('should pass a basic assertion', () => {
    expect(1 + 1).toBe(2);
  });
});
