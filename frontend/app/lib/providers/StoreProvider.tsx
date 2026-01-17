"use client";

import { useEffect, useRef, type ReactNode } from "react";
import { useAppStore } from "../store";

interface StoreProviderProps {
  children: ReactNode;
}

/**
 * StoreProvider initializes the store on mount.
 * It checks authentication and loads projects from localStorage.
 */
export function StoreProvider({ children }: StoreProviderProps) {
  const initialized = useRef(false);
  const loadProjects = useAppStore((state) => state.loadProjects);
  const checkAuth = useAppStore((state) => state.checkAuth);

  useEffect(() => {
    if (!initialized.current) {
      initialized.current = true;
      // Check authentication first, then load projects
      checkAuth().then(() => {
        loadProjects();
      });
    }
  }, [loadProjects, checkAuth]);

  return <>{children}</>;
}
