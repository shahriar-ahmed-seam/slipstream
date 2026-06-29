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
        <div className="mx-auto max-w-7xl px-8 py-16">
          <div className="border border-rose-200 bg-rose-50 p-6">
            <h3 className="text-base font-semibold text-rose-700">
              Something went wrong rendering this panel.
            </h3>
            <p className="mt-2 text-sm text-rose-600">{this.state.message}</p>
            <button
              type="button"
              onClick={this.reset}
              className="mt-4 border border-rose-300 bg-white px-4 py-1.5 text-xs font-medium text-rose-700 transition hover:bg-rose-100"
            >
              Try again
            </button>
          </div>
        </div>
      );
    }
    return this.props.children;
  }
}