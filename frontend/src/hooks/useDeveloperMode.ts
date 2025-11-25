import { useState, useEffect, useCallback } from 'react';

const STORAGE_KEY = 'mtga-companion-developer-mode';
const CLICK_COUNT_THRESHOLD = 5;
const CLICK_TIMEOUT_MS = 3000;

export function useDeveloperMode() {
  const [isDeveloperMode, setIsDeveloperMode] = useState(() => {
    // Initialize from localStorage
    if (typeof window !== 'undefined') {
      return localStorage.getItem(STORAGE_KEY) === 'true';
    }
    return false;
  });

  const [clickCount, setClickCount] = useState(0);
  const [lastClickTime, setLastClickTime] = useState(0);

  // Persist to localStorage when changed
  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, String(isDeveloperMode));
  }, [isDeveloperMode]);

  // Handle version click for secret activation
  const handleVersionClick = useCallback(() => {
    const now = Date.now();

    // Reset if too much time has passed
    if (now - lastClickTime > CLICK_TIMEOUT_MS) {
      setClickCount(1);
      setLastClickTime(now);
      return;
    }

    const newCount = clickCount + 1;
    setClickCount(newCount);
    setLastClickTime(now);

    // Toggle developer mode after threshold clicks
    if (newCount >= CLICK_COUNT_THRESHOLD) {
      setIsDeveloperMode((prev) => !prev);
      setClickCount(0);
    }
  }, [clickCount, lastClickTime]);

  // Direct toggle for explicit enable/disable
  const toggleDeveloperMode = useCallback(() => {
    setIsDeveloperMode((prev) => !prev);
  }, []);

  return {
    isDeveloperMode,
    handleVersionClick,
    toggleDeveloperMode,
    clickCount, // Exposed for UI feedback if needed
  };
}
