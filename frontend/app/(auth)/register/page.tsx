"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAppStore } from "../../lib/store";

export default function RegisterPage() {
  const router = useRouter();
  const register = useAppStore((state) => state.register);
  const isAuthenticated = useAppStore((state) => state.isAuthenticated);
  const isAuthLoading = useAppStore((state) => state.isAuthLoading);
  const authError = useAppStore((state) => state.authError);
  const clearAuthError = useAppStore((state) => state.clearAuthError);
  const checkAuth = useAppStore((state) => state.checkAuth);

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [tenantName, setTenantName] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [validationError, setValidationError] = useState<string | null>(null);

  // Check if already authenticated
  useEffect(() => {
    checkAuth();
  }, [checkAuth]);

  useEffect(() => {
    if (isAuthenticated && !isAuthLoading) {
      router.push("/dashboard");
    }
  }, [isAuthenticated, isAuthLoading, router]);

  // Clear errors when inputs change
  const clearErrors = useCallback(() => {
    if (authError) clearAuthError();
    if (validationError) setValidationError(null);
  }, [authError, clearAuthError, validationError]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Validate inputs
    if (!email || !password || !confirmPassword || !tenantName) {
      setValidationError("All fields are required");
      return;
    }

    if (password.length < 8) {
      setValidationError("Password must be at least 8 characters");
      return;
    }

    if (password !== confirmPassword) {
      setValidationError("Passwords do not match");
      return;
    }

    setIsSubmitting(true);
    const success = await register({ email, password, tenant_name: tenantName });
    setIsSubmitting(false);

    if (success) {
      router.push("/dashboard");
    }
  };

  // Show loading spinner while checking auth
  if (isAuthLoading) {
    return (
      <div className="min-h-screen bg-[var(--obsidian)] flex items-center justify-center">
        <div className="spinner" />
      </div>
    );
  }

  const error = validationError || authError;

  return (
    <main className="min-h-screen bg-[var(--obsidian)] relative overflow-hidden">
      {/* Background gradient mesh */}
      <div className="gradient-mesh">
        <div className="gradient-orb gradient-orb-1" style={{ opacity: 0.4 }} />
        <div className="gradient-orb gradient-orb-2" style={{ opacity: 0.3 }} />
        <div className="grid-overlay" />
        <div className="noise-overlay" />
      </div>

      {/* Navigation */}
      <nav className="relative z-10 p-6">
        <Link href="/" className="nav-logo inline-flex">
          <div className="nav-logo-icon">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
              <path d="M12 2L2 7l10 5 10-5-10-5z" />
              <path d="M2 17l10 5 10-5" />
              <path d="M2 12l10 5 10-5" />
            </svg>
          </div>
          <span>Sovereign Firm</span>
        </Link>
      </nav>

      {/* Register Form */}
      <div className="relative z-10 flex items-center justify-center min-h-[calc(100vh-100px)] py-8">
        <div className="w-full max-w-md px-6">
          <div className="glass-card rounded-2xl p-8 animate-scale-in">
            {/* Header */}
            <div className="text-center mb-8">
              <h1 className="text-display-sm text-gradient-white mb-3">Create Account</h1>
              <p className="text-[var(--silver)]">
                Start building AI-powered software in minutes
              </p>
            </div>

            {/* Error message */}
            {error && (
              <div className="mb-6 p-4 rounded-lg bg-[rgba(244,63,94,0.1)] border border-[var(--rose-glow)] text-[var(--rose-glow)] text-sm animate-slide-down">
                {error}
              </div>
            )}

            {/* Form */}
            <form onSubmit={handleSubmit} className="space-y-5">
              <div>
                <label htmlFor="tenantName" className="input-label">
                  Company / Organization Name
                </label>
                <input
                  id="tenantName"
                  type="text"
                  value={tenantName}
                  onChange={(e) => { setTenantName(e.target.value); clearErrors(); }}
                  className="input"
                  placeholder="Acme Inc."
                  required
                  autoFocus
                />
              </div>

              <div>
                <label htmlFor="email" className="input-label">
                  Email Address
                </label>
                <input
                  id="email"
                  type="email"
                  value={email}
                  onChange={(e) => { setEmail(e.target.value); clearErrors(); }}
                  className="input"
                  placeholder="you@company.com"
                  required
                  autoComplete="email"
                />
              </div>

              <div>
                <label htmlFor="password" className="input-label">
                  Password
                </label>
                <input
                  id="password"
                  type="password"
                  value={password}
                  onChange={(e) => { setPassword(e.target.value); clearErrors(); }}
                  className="input"
                  placeholder="Minimum 8 characters"
                  required
                  autoComplete="new-password"
                  minLength={8}
                />
              </div>

              <div>
                <label htmlFor="confirmPassword" className="input-label">
                  Confirm Password
                </label>
                <input
                  id="confirmPassword"
                  type="password"
                  value={confirmPassword}
                  onChange={(e) => { setConfirmPassword(e.target.value); clearErrors(); }}
                  className="input"
                  placeholder="Re-enter your password"
                  required
                  autoComplete="new-password"
                />
              </div>

              <button
                type="submit"
                disabled={isSubmitting || !email || !password || !confirmPassword || !tenantName}
                className="btn btn-primary w-full py-4 text-base disabled:opacity-50 disabled:cursor-not-allowed mt-2"
              >
                {isSubmitting ? (
                  <>
                    <div className="spinner w-5 h-5 border-2 border-[var(--void)] border-t-transparent" />
                    <span>Creating account...</span>
                  </>
                ) : (
                  <>
                    <span>Create Account</span>
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <path d="M5 12h14M12 5l7 7-7 7" />
                    </svg>
                  </>
                )}
              </button>
            </form>

            {/* Divider */}
            <div className="my-8 flex items-center gap-4">
              <div className="flex-1 h-px bg-[var(--steel)]" />
              <span className="text-sm text-[var(--silver)]">or</span>
              <div className="flex-1 h-px bg-[var(--steel)]" />
            </div>

            {/* Login link */}
            <div className="text-center">
              <p className="text-[var(--silver)] text-sm">
                Already have an account?{" "}
                <Link
                  href="/login"
                  className="text-[var(--cyan-glow)] hover:underline font-medium"
                >
                  Sign in
                </Link>
              </p>
            </div>
          </div>

          {/* Footer text */}
          <p className="text-center text-xs text-[var(--silver)] mt-8">
            By creating an account, you agree to our{" "}
            <a href="#" className="text-[var(--pearl)] hover:underline">Terms of Service</a>
            {" "}and{" "}
            <a href="#" className="text-[var(--pearl)] hover:underline">Privacy Policy</a>
          </p>
        </div>
      </div>

      {/* Floating particles */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        {[
          { left: 10, top: 20, delay: 0 },
          { left: 85, top: 15, delay: 1.2 },
          { left: 25, top: 70, delay: 0.5 },
          { left: 70, top: 85, delay: 2.1 },
          { left: 45, top: 30, delay: 0.8 },
        ].map((p, i) => (
          <div
            key={i}
            className="absolute w-1 h-1 bg-[var(--violet-glow)] rounded-full opacity-30 animate-float"
            style={{
              left: `${p.left}%`,
              top: `${p.top}%`,
              animationDelay: `${p.delay}s`,
              animationDuration: "8s",
            }}
          />
        ))}
      </div>
    </main>
  );
}
