import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import SettingItem from './SettingItem';

describe('SettingItem', () => {
  describe('Basic Rendering', () => {
    it('should render label text', () => {
      render(
        <SettingItem label="Test Label">
          <button>Test Button</button>
        </SettingItem>
      );

      expect(screen.getByText('Test Label')).toBeInTheDocument();
    });

    it('should render children', () => {
      render(
        <SettingItem label="Test Label">
          <button>Test Button</button>
        </SettingItem>
      );

      expect(screen.getByRole('button', { name: 'Test Button' })).toBeInTheDocument();
    });

    it('should render description when provided', () => {
      render(
        <SettingItem label="Test Label" description="Test Description">
          <button>Test Button</button>
        </SettingItem>
      );

      expect(screen.getByText('Test Description')).toBeInTheDocument();
    });

    it('should not render description when not provided', () => {
      render(
        <SettingItem label="Test Label">
          <button>Test Button</button>
        </SettingItem>
      );

      expect(document.querySelector('.setting-description')).not.toBeInTheDocument();
    });

    it('should render hint when provided', () => {
      render(
        <SettingItem label="Test Label" hint="ws://localhost:9999">
          <input type="number" />
        </SettingItem>
      );

      expect(screen.getByText('ws://localhost:9999')).toBeInTheDocument();
    });

    it('should not render hint when not provided', () => {
      render(
        <SettingItem label="Test Label">
          <input type="number" />
        </SettingItem>
      );

      expect(document.querySelector('.setting-hint')).not.toBeInTheDocument();
    });
  });

  describe('CSS Classes', () => {
    it('should have setting-item class by default', () => {
      render(
        <SettingItem label="Test Label">
          <button>Test Button</button>
        </SettingItem>
      );

      const container = document.querySelector('.setting-item');
      expect(container).toBeInTheDocument();
    });

    it('should add indented class when indented prop is true', () => {
      render(
        <SettingItem label="Test Label" indented>
          <button>Test Button</button>
        </SettingItem>
      );

      const container = document.querySelector('.setting-item');
      expect(container).toHaveClass('indented');
    });

    it('should not have indented class when indented prop is false', () => {
      render(
        <SettingItem label="Test Label" indented={false}>
          <button>Test Button</button>
        </SettingItem>
      );

      const container = document.querySelector('.setting-item');
      expect(container).not.toHaveClass('indented');
    });

    it('should add danger class when danger prop is true', () => {
      render(
        <SettingItem label="Test Label" danger>
          <button>Delete</button>
        </SettingItem>
      );

      const container = document.querySelector('.setting-item');
      expect(container).toHaveClass('danger');
    });

    it('should not have danger class when danger prop is false', () => {
      render(
        <SettingItem label="Test Label" danger={false}>
          <button>Test Button</button>
        </SettingItem>
      );

      const container = document.querySelector('.setting-item');
      expect(container).not.toHaveClass('danger');
    });

    it('should have both indented and danger classes when both props are true', () => {
      render(
        <SettingItem label="Test Label" indented danger>
          <button>Delete</button>
        </SettingItem>
      );

      const container = document.querySelector('.setting-item');
      expect(container).toHaveClass('indented');
      expect(container).toHaveClass('danger');
    });
  });

  describe('Structure', () => {
    it('should have setting-label element', () => {
      render(
        <SettingItem label="Test Label">
          <button>Test Button</button>
        </SettingItem>
      );

      expect(document.querySelector('.setting-label')).toBeInTheDocument();
    });

    it('should have setting-control element', () => {
      render(
        <SettingItem label="Test Label">
          <button>Test Button</button>
        </SettingItem>
      );

      expect(document.querySelector('.setting-control')).toBeInTheDocument();
    });

    it('should wrap description in setting-description span', () => {
      render(
        <SettingItem label="Test Label" description="Test Description">
          <button>Test Button</button>
        </SettingItem>
      );

      const description = screen.getByText('Test Description');
      expect(description.tagName).toBe('SPAN');
      expect(description).toHaveClass('setting-description');
    });

    it('should wrap hint in setting-hint span', () => {
      render(
        <SettingItem label="Test Label" hint="Test Hint">
          <button>Test Button</button>
        </SettingItem>
      );

      const hint = screen.getByText('Test Hint');
      expect(hint.tagName).toBe('SPAN');
      expect(hint).toHaveClass('setting-hint');
    });
  });

  describe('Multiple Children', () => {
    it('should render multiple children', () => {
      render(
        <SettingItem label="Test Label">
          <button>Button 1</button>
          <button>Button 2</button>
        </SettingItem>
      );

      expect(screen.getByRole('button', { name: 'Button 1' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Button 2' })).toBeInTheDocument();
    });
  });
});
