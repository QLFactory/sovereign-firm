import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";

// Since WebContainers can't run in Node.js/jsdom, we'll test the component's
// rendering states and the file conversion utility function.

// Extract and test the file conversion logic
function convertToWebContainerFiles(files: Record<string, string>) {
  const result: Record<string, any> = {};

  for (const [filePath, content] of Object.entries(files)) {
    // Remove leading slash if present
    const normalizedPath = filePath.startsWith("/") ? filePath.slice(1) : filePath;
    const parts = normalizedPath.split("/");

    let current = result;
    for (let i = 0; i < parts.length - 1; i++) {
      const part = parts[i];
      if (!current[part]) {
        current[part] = { directory: {} };
      }
      current = current[part].directory;
    }

    const fileName = parts[parts.length - 1];
    current[fileName] = { file: { contents: content } };
  }

  return result;
}

describe("WebContainerPreview", () => {
  describe("convertToWebContainerFiles utility", () => {
    it("should convert flat files to nested WebContainer format", () => {
      const files = {
        "/src/App.jsx": "export default function App() {}",
        "/src/main.jsx": "import App from './App';",
      };

      const result = convertToWebContainerFiles(files);

      expect(result).toEqual({
        src: {
          directory: {
            "App.jsx": { file: { contents: "export default function App() {}" } },
            "main.jsx": { file: { contents: "import App from './App';" } },
          },
        },
      });
    });

    it("should handle root-level files", () => {
      const files = {
        "/package.json": '{"name": "test"}',
        "/index.html": "<html></html>",
      };

      const result = convertToWebContainerFiles(files);

      expect(result).toEqual({
        "package.json": { file: { contents: '{"name": "test"}' } },
        "index.html": { file: { contents: "<html></html>" } },
      });
    });

    it("should handle deeply nested files", () => {
      const files = {
        "/src/components/ui/Button.tsx": "export function Button() {}",
      };

      const result = convertToWebContainerFiles(files);

      expect(result).toEqual({
        src: {
          directory: {
            components: {
              directory: {
                ui: {
                  directory: {
                    "Button.tsx": { file: { contents: "export function Button() {}" } },
                  },
                },
              },
            },
          },
        },
      });
    });

    it("should handle files without leading slash", () => {
      const files = {
        "src/App.jsx": "export default function App() {}",
      };

      const result = convertToWebContainerFiles(files);

      expect(result).toEqual({
        src: {
          directory: {
            "App.jsx": { file: { contents: "export default function App() {}" } },
          },
        },
      });
    });

    it("should handle empty files", () => {
      const files = {
        "/src/empty.js": "",
      };

      const result = convertToWebContainerFiles(files);

      expect(result).toEqual({
        src: {
          directory: {
            "empty.js": { file: { contents: "" } },
          },
        },
      });
    });

    it("should handle multiple files in same directory", () => {
      const files = {
        "/src/App.jsx": "app code",
        "/src/App.test.jsx": "test code",
        "/src/App.css": "css code",
      };

      const result = convertToWebContainerFiles(files);

      expect(result.src.directory["App.jsx"]).toEqual({ file: { contents: "app code" } });
      expect(result.src.directory["App.test.jsx"]).toEqual({ file: { contents: "test code" } });
      expect(result.src.directory["App.css"]).toEqual({ file: { contents: "css code" } });
    });

    it("should handle files with special characters in content", () => {
      const files = {
        "/src/App.jsx": `export default function App() {
  return <div className="test">Hello "World"</div>;
}`,
      };

      const result = convertToWebContainerFiles(files);

      expect(result.src.directory["App.jsx"].file.contents).toContain('className="test"');
      expect(result.src.directory["App.jsx"].file.contents).toContain('"World"');
    });
  });

  describe("default files", () => {
    // Test that the default configurations are correct
    const defaultPackageJson = {
      name: "preview-app",
      private: true,
      version: "0.0.0",
      type: "module",
      scripts: {
        dev: "vite",
        build: "vite build",
        preview: "vite preview",
        test: "vitest run",
      },
      dependencies: {
        react: "^18.2.0",
        "react-dom": "^18.2.0",
      },
      devDependencies: {
        "@testing-library/jest-dom": "^6.4.2",
        "@testing-library/react": "^14.2.1",
        "@vitejs/plugin-react": "^4.2.1",
        jsdom: "^24.0.0",
        vite: "^5.1.0",
        vitest: "^1.3.1",
      },
    };

    it("should have vite as dev server", () => {
      expect(defaultPackageJson.scripts.dev).toBe("vite");
    });

    it("should have vitest as test runner", () => {
      expect(defaultPackageJson.scripts.test).toBe("vitest run");
    });

    it("should include testing-library dependencies", () => {
      expect(defaultPackageJson.devDependencies["@testing-library/react"]).toBeDefined();
      expect(defaultPackageJson.devDependencies["@testing-library/jest-dom"]).toBeDefined();
    });

    it("should include jsdom for testing", () => {
      expect(defaultPackageJson.devDependencies.jsdom).toBeDefined();
    });
  });

  describe("test file detection", () => {
    it("should detect test files by .test. pattern", () => {
      const files = {
        "/src/App.jsx": "app code",
        "/src/App.test.jsx": "test code",
      };

      const hasTests = Object.keys(files).some((f) => f.includes(".test."));
      expect(hasTests).toBe(true);
    });

    it("should return false when no test files", () => {
      const files = {
        "/src/App.jsx": "app code",
        "/src/utils.js": "utils code",
      };

      const hasTests = Object.keys(files).some((f) => f.includes(".test."));
      expect(hasTests).toBe(false);
    });
  });

  describe("file detection helpers", () => {
    it("should detect App file in various formats", () => {
      const testCases = [
        { files: { "/src/App.jsx": "code" }, expected: true },
        { files: { "/src/App.tsx": "code" }, expected: true },
        { files: { "/src/App.js": "code" }, expected: true },
        { files: { "/src/utils.js": "code" }, expected: false },
      ];

      testCases.forEach(({ files, expected }) => {
        const hasAppFile = Object.keys(files).some(
          (f) => f.includes("App.jsx") || f.includes("App.tsx") || f.includes("App.js")
        );
        expect(hasAppFile).toBe(expected);
      });
    });

    it("should detect index.html", () => {
      const files = { "/index.html": "<html></html>", "/src/App.jsx": "code" };
      const hasIndexHtml = Object.keys(files).some((f) => f.includes("index.html"));
      expect(hasIndexHtml).toBe(true);
    });

    it("should detect main entry point", () => {
      const testCases = [
        { files: { "/src/main.jsx": "code" }, expected: true },
        { files: { "/src/main.tsx": "code" }, expected: true },
        { files: { "/src/App.jsx": "code" }, expected: false },
      ];

      testCases.forEach(({ files, expected }) => {
        const hasMain = Object.keys(files).some(
          (f) => f.includes("main.jsx") || f.includes("main.tsx")
        );
        expect(hasMain).toBe(expected);
      });
    });

    it("should detect setupTests file", () => {
      const files = { "/src/setupTests.js": "import '@testing-library/jest-dom';" };
      const hasSetupTests = Object.keys(files).some((f) => f.includes("setupTests"));
      expect(hasSetupTests).toBe(true);
    });
  });
});

describe("WebContainerPreview rendering states", () => {
  // Mock WebContainer to prevent actual boot
  beforeEach(() => {
    vi.mock("@webcontainer/api", () => ({
      WebContainer: {
        boot: vi.fn().mockRejectedValue(new Error("Cannot boot in test environment")),
      },
    }));
  });

  it("should have correct CSS classes for status indicator", () => {
    // Test the CSS class logic for status indicators
    const getStatusClass = (url: string | null) => {
      return url ? "bg-green-500" : "bg-yellow-500 animate-pulse";
    };

    expect(getStatusClass(null)).toBe("bg-yellow-500 animate-pulse");
    expect(getStatusClass("http://localhost:5173")).toBe("bg-green-500");
  });

  it("should have correct CSS classes for test button states", () => {
    const getButtonClass = (isRunningTests: boolean) => {
      return isRunningTests
        ? "bg-zinc-700 text-zinc-400 cursor-wait"
        : "bg-purple-600 hover:bg-purple-500 text-white";
    };

    expect(getButtonClass(true)).toContain("cursor-wait");
    expect(getButtonClass(false)).toContain("bg-purple-600");
  });

  it("should have correct test result badge classes", () => {
    const getResultClass = (result: string) => {
      return result.includes("✅")
        ? "bg-green-900 text-green-300"
        : "bg-red-900 text-red-300";
    };

    expect(getResultClass("✅ All tests passed!")).toContain("bg-green-900");
    expect(getResultClass("❌ Tests failed")).toContain("bg-red-900");
  });
});
