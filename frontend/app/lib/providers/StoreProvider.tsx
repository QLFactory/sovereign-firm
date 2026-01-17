"use client";

import { useEffect, useRef, type ReactNode } from "react";
import { useAppStore } from "../store";

interface StoreProviderProps {
  children: ReactNode;
}

/**
 * StoreProvider initializes the store on mount.
 * It loads projects from localStorage and sets up any initial state.
 */
export function StoreProvider({ children }: StoreProviderProps) {
  const initialized = useRef(false);
  const loadProjects = useAppStore((state) => state.loadProjects);

  useEffect(() => {
    if (!initialized.current) {
      initialized.current = true;
      loadProjects();
    }
  }, [loadProjects]);

  return <>{children}</>;
}
