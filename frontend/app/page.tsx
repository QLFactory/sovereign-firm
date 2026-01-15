import { Suspense } from "react";
import PodConsole from "./components/PodConsole";
import ErrorBoundary from "./components/ErrorBoundary";

function LoadingFallback() {
  return (
    <div className="min-h-screen bg-black flex items-center justify-center">
      <div className="text-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
        <p className="text-zinc-400">Loading Sovereign Firm...</p>
      </div>
    </div>
  );
}

export default function Home() {
  return (
    <main className="min-h-screen bg-black">
      <ErrorBoundary>
        <Suspense fallback={<LoadingFallback />}>
          <PodConsole />
        </Suspense>
      </ErrorBoundary>
    </main>
  );
}
