"use client";

import { Component, type ReactNode } from "react";

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  message: string;
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false, message: "" };

  static getDerivedStateFromError(error: unknown): State {
    const message =
      error instanceof Error ? error.message : "Unknown rendering failure.";
    return { hasError: true, message };
  }

  componentDidCatch(error: unknown): void {
    // In production, this would forward to an error tracker.
    // eslint-disable-next-line no-console
    console.error("[ErrorBoundary]", error);
  }

  reset = (): void => this.setState({ hasError: false, message: "" });

  render(): ReactNode {
    if (this.state.hasError) {
      if (this.props.fallback) return this.props.fallback;
      return (
        <div className="surface p-6 text-rose-300">
          <h3 className="text-base font-semibold text-rose-200">
            Something went wrong rendering this panel.
          </h3>
          <p className="mt-2 text-sm text-rose-300/80">{this.state.message}</p>
          <button
            type="button"
            onClick={this.reset}
            className="mt-4 rounded-md border border-rose-500/40 px-3 py-1 text-xs text-rose-200 hover:bg-rose-500/10"
          >
            Try again
          </button>
        </div>
      );
    }
    return this.props.children;
  }
}