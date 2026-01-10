import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import HelpIcon from './HelpIcon';

describe('HelpIcon', () => {
  describe('rendering', () => {
    it('renders a help button', () => {
      render(
        <HelpIcon title="Test Help">
          <p>Help content</p>
        </HelpIcon>
      );

      expect(screen.getByRole('button', { name: /help: test help/i })).toBeInTheDocument();
    });

    it('shows ? character in button', () => {
      render(
        <HelpIcon title="Test">
          <p>Content</p>
        </HelpIcon>
      );

      expect(screen.getByRole('button')).toHaveTextContent('?');
    });

    it('does not show popover initially', () => {
      render(
        <HelpIcon title="Test">
          <p>Help content</p>
        </HelpIcon>
      );

      expect(screen.queryByText('Help content')).not.toBeInTheDocument();
    });
  });

  describe('popover interaction', () => {
    it('shows popover when clicked', async () => {
      const user = userEvent.setup();
      render(
        <HelpIcon title="Test Title">
          <p>Help content here</p>
        </HelpIcon>
      );

      await user.click(screen.getByRole('button', { name: /help/i }));

      expect(screen.getByText('Test Title')).toBeInTheDocument();
      expect(screen.getByText('Help content here')).toBeInTheDocument();
    });

    it('hides popover when clicked again', async () => {
      const user = userEvent.setup();
      render(
        <HelpIcon title="Test">
          <p>Content</p>
        </HelpIcon>
      );

      const button = screen.getByRole('button', { name: /help/i });
      await user.click(button);
      expect(screen.getByText('Content')).toBeInTheDocument();

      await user.click(button);
      expect(screen.queryByText('Content')).not.toBeInTheDocument();
    });

    it('closes popover when close button is clicked', async () => {
      const user = userEvent.setup();
      render(
        <HelpIcon title="Test">
          <p>Content</p>
        </HelpIcon>
      );

      await user.click(screen.getByRole('button', { name: /help/i }));
      expect(screen.getByText('Content')).toBeInTheDocument();

      await user.click(screen.getByRole('button', { name: /close/i }));
      expect(screen.queryByText('Content')).not.toBeInTheDocument();
    });

    it('closes popover on escape key', async () => {
      const user = userEvent.setup();
      render(
        <HelpIcon title="Test">
          <p>Content</p>
        </HelpIcon>
      );

      await user.click(screen.getByRole('button', { name: /help/i }));
      expect(screen.getByText('Content')).toBeInTheDocument();

      await user.keyboard('{Escape}');
      expect(screen.queryByText('Content')).not.toBeInTheDocument();
    });

    it('closes popover when clicking outside', async () => {
      const user = userEvent.setup();
      render(
        <div>
          <HelpIcon title="Test">
            <p>Content</p>
          </HelpIcon>
          <button>Outside button</button>
        </div>
      );

      await user.click(screen.getByRole('button', { name: /help/i }));
      expect(screen.getByText('Content')).toBeInTheDocument();

      await user.click(screen.getByRole('button', { name: /outside/i }));
      expect(screen.queryByText('Content')).not.toBeInTheDocument();
    });
  });

  describe('accessibility', () => {
    it('has aria-expanded attribute', async () => {
      const user = userEvent.setup();
      render(
        <HelpIcon title="Test">
          <p>Content</p>
        </HelpIcon>
      );

      const button = screen.getByRole('button', { name: /help/i });
      expect(button).toHaveAttribute('aria-expanded', 'false');

      await user.click(button);
      expect(button).toHaveAttribute('aria-expanded', 'true');
    });

    it('has aria-label with title', () => {
      render(
        <HelpIcon title="My Feature">
          <p>Content</p>
        </HelpIcon>
      );

      expect(screen.getByRole('button')).toHaveAttribute('aria-label', 'Help: My Feature');
    });
  });

  describe('size variants', () => {
    it('applies small size class by default', () => {
      render(
        <HelpIcon title="Test">
          <p>Content</p>
        </HelpIcon>
      );

      expect(screen.getByRole('button', { name: /help/i })).toHaveClass('help-icon-small');
    });

    it('applies medium size class', () => {
      render(
        <HelpIcon title="Test" size="medium">
          <p>Content</p>
        </HelpIcon>
      );

      expect(screen.getByRole('button', { name: /help/i })).toHaveClass('help-icon-medium');
    });

    it('applies large size class', () => {
      render(
        <HelpIcon title="Test" size="large">
          <p>Content</p>
        </HelpIcon>
      );

      expect(screen.getByRole('button', { name: /help/i })).toHaveClass('help-icon-large');
    });
  });

  describe('position variants', () => {
    it('applies bottom position by default', async () => {
      const user = userEvent.setup();
      render(
        <HelpIcon title="Test">
          <p>Content</p>
        </HelpIcon>
      );

      await user.click(screen.getByRole('button', { name: /help/i }));
      expect(document.querySelector('.help-popover')).toHaveClass('help-popover-bottom');
    });

    it('applies top position', async () => {
      const user = userEvent.setup();
      render(
        <HelpIcon title="Test" position="top">
          <p>Content</p>
        </HelpIcon>
      );

      await user.click(screen.getByRole('button', { name: /help/i }));
      expect(document.querySelector('.help-popover')).toHaveClass('help-popover-top');
    });

    it('applies left position', async () => {
      const user = userEvent.setup();
      render(
        <HelpIcon title="Test" position="left">
          <p>Content</p>
        </HelpIcon>
      );

      await user.click(screen.getByRole('button', { name: /help/i }));
      expect(document.querySelector('.help-popover')).toHaveClass('help-popover-left');
    });

    it('applies right position', async () => {
      const user = userEvent.setup();
      render(
        <HelpIcon title="Test" position="right">
          <p>Content</p>
        </HelpIcon>
      );

      await user.click(screen.getByRole('button', { name: /help/i }));
      expect(document.querySelector('.help-popover')).toHaveClass('help-popover-right');
    });
  });

  describe('content rendering', () => {
    it('renders complex content', async () => {
      const user = userEvent.setup();
      render(
        <HelpIcon title="Complex Help">
          <p>First paragraph</p>
          <ul>
            <li>Item 1</li>
            <li>Item 2</li>
          </ul>
          <p>
            <strong>Bold text</strong> and <code>code</code>
          </p>
        </HelpIcon>
      );

      await user.click(screen.getByRole('button', { name: /help/i }));

      expect(screen.getByText('First paragraph')).toBeInTheDocument();
      expect(screen.getByText('Item 1')).toBeInTheDocument();
      expect(screen.getByText('Item 2')).toBeInTheDocument();
      expect(screen.getByText('Bold text')).toBeInTheDocument();
      expect(screen.getByText('code')).toBeInTheDocument();
    });
  });
});
