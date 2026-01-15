"use client";

import React, { Component, ErrorInfo, ReactNode } from "react";

interface Props {
    children: ReactNode;
    fallback?: ReactNode;
}

interface State {
    hasError: boolean;
    error: Error | null;
    errorInfo: ErrorInfo | null;
}

class ErrorBoundary extends Component<Props, State> {
    public state: State = {
        hasError: false,
        error: null,
        errorInfo: null,
    };

    public static getDerivedStateFromError(error: Error): State {
        return { hasError: true, error, errorInfo: null };
    }

    public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
        console.error("Error boundary caught error:", error, errorInfo);
        this.setState({ errorInfo });
    }

    private handleRetry = () => {
        this.setState({ hasError: false, error: null, errorInfo: null });
    };

    public render() {
        if (this.state.hasError) {
            if (this.props.fallback) {
                return this.props.fallback;
            }

            return (
                <div className="min-h-screen bg-zinc-950 flex items-center justify-center p-4">
                    <div className="max-w-md w-full bg-zinc-900 rounded-lg p-6 border border-zinc-800">
                        <div className="flex items-center gap-3 mb-4">
                            <div className="w-10 h-10 bg-red-500/20 rounded-full flex items-center justify-center">
                                <svg
                                    className="w-6 h-6 text-red-500"
                                    fill="none"
                                    viewBox="0 0 24 24"
                                    stroke="currentColor"
                                >
                                    <path
                                        strokeLinecap="round"
                                        strokeLinejoin="round"
                                        strokeWidth={2}
                                        d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                                    />
                                </svg>
                            </div>
                            <div>
                                <h2 className="text-lg font-semibold text-white">
                                    Something went wrong
                                </h2>
                                <p className="text-sm text-zinc-400">
                                    An error occurred while rendering
                                </p>
                            </div>
                        </div>

                        {this.state.error && (
                            <div className="bg-zinc-950 rounded p-3 mb-4 overflow-x-auto">
                                <code className="text-sm text-red-400 font-mono">
                                    {this.state.error.message}
                                </code>
                            </div>
                        )}

                        <div className="flex gap-2">
                            <button
                                onClick={this.handleRetry}
                                className="flex-1 bg-blue-600 hover:bg-blue-500 text-white font-medium py-2 px-4 rounded transition-colors"
                            >
                                Try Again
                            </button>
                            <button
                                onClick={() => window.location.reload()}
                                className="flex-1 bg-zinc-800 hover:bg-zinc-700 text-white font-medium py-2 px-4 rounded transition-colors"
                            >
                                Reload Page
                            </button>
                        </div>

                        {process.env.NODE_ENV === "development" && this.state.errorInfo && (
                            <details className="mt-4">
                                <summary className="text-sm text-zinc-400 cursor-pointer hover:text-zinc-300">
                                    Show stack trace
                                </summary>
                                <pre className="mt-2 p-3 bg-zinc-950 rounded text-xs text-zinc-500 overflow-x-auto">
                                    {this.state.errorInfo.componentStack}
                                </pre>
                            </details>
                        )}
                    </div>
                </div>
            );
        }

        return this.props.children;
    }
}

export default ErrorBoundary;
