import PodConsole from "./components/PodConsole";
import ErrorBoundary from "./components/ErrorBoundary";

export default function Home() {
  return (
    <main className="min-h-screen bg-black">
      <ErrorBoundary>
        <PodConsole />
      </ErrorBoundary>
    </main>
  );
}
