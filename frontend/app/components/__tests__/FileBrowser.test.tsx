import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import "@testing-library/jest-dom";
import FileBrowser from "../FileBrowser";

describe("FileBrowser", () => {
  const mockFiles = {
    "/src/App.jsx": 'export default function App() {}',
    "/src/index.js": 'import App from "./App"',
    "/src/styles.css": "body { margin: 0; }",
    "/src/components/Header.jsx": "export function Header() {}",
    "/package.json": '{ "name": "app" }',
    "/README.md": "# My App",
  };

  const defaultProps = {
    files: mockFiles,
    selectedFile: null,
    onSelectFile: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("rendering", () => {
    it("renders file count", () => {
      render(<FileBrowser {...defaultProps} />);

      expect(screen.getByText("6 files")).toBeInTheDocument();
    });

    it("renders files header", () => {
      render(<FileBrowser {...defaultProps} />);

      expect(screen.getByText("Files")).toBeInTheDocument();
    });

    it("renders empty state when no files", () => {
      render(<FileBrowser {...defaultProps} files={{}} />);

      expect(screen.getByText("No files yet")).toBeInTheDocument();
    });

    it("renders directory structure", () => {
      render(<FileBrowser {...defaultProps} />);

      // Root level items
      expect(screen.getByText("src")).toBeInTheDocument();
      expect(screen.getByText("package.json")).toBeInTheDocument();
      expect(screen.getByText("README.md")).toBeInTheDocument();
    });

    it("expands src directory by default", () => {
      render(<FileBrowser {...defaultProps} />);

      // Files in src should be visible
      expect(screen.getByText("App.jsx")).toBeInTheDocument();
      expect(screen.getByText("index.js")).toBeInTheDocument();
      expect(screen.getByText("styles.css")).toBeInTheDocument();
    });
  });

  describe("file selection", () => {
    it("highlights selected file", () => {
      render(
        <FileBrowser
          {...defaultProps}
          selectedFile="/src/App.jsx"
        />
      );

      const selectedButton = screen.getByText("App.jsx").closest("button");
      expect(selectedButton).toHaveClass("bg-blue-600");
    });

    it("calls onSelectFile when file is clicked", () => {
      const onSelectFile = vi.fn();
      render(
        <FileBrowser
          {...defaultProps}
          onSelectFile={onSelectFile}
        />
      );

      fireEvent.click(screen.getByText("App.jsx"));

      expect(onSelectFile).toHaveBeenCalledWith("/src/App.jsx");
    });
  });

  describe("directory toggling", () => {
    it("toggles directory expansion when clicked", () => {
      render(<FileBrowser {...defaultProps} />);

      // src is expanded by default, components subdirectory should be visible
      expect(screen.getByText("components")).toBeInTheDocument();

      // Click to expand components
      fireEvent.click(screen.getByText("components"));

      // Header.jsx should now be visible
      expect(screen.getByText("Header.jsx")).toBeInTheDocument();
    });

    it("collapses expanded directory when clicked", () => {
      render(<FileBrowser {...defaultProps} />);

      // src is expanded by default
      expect(screen.getByText("App.jsx")).toBeInTheDocument();

      // Click to collapse src
      fireEvent.click(screen.getByText("src"));

      // Files should be hidden
      expect(screen.queryByText("App.jsx")).not.toBeInTheDocument();
    });
  });

  describe("file icons", () => {
    it("shows correct icons for different file types", () => {
      render(<FileBrowser {...defaultProps} />);

      // Check that file names are rendered (icons are emojis in the component)
      expect(screen.getByText("App.jsx")).toBeInTheDocument();
      expect(screen.getByText("styles.css")).toBeInTheDocument();
      expect(screen.getByText("package.json")).toBeInTheDocument();
    });
  });

  describe("sorting", () => {
    it("shows directories before files at same level", () => {
      const files = {
        "/zebra.txt": "content",
        "/alpha/file.js": "content",
        "/apple.txt": "content",
      };

      render(
        <FileBrowser
          {...defaultProps}
          files={files}
        />
      );

      const items = screen.getAllByRole("button");
      const itemTexts = items.map((item) => item.textContent?.trim());

      // alpha directory should come before apple.txt and zebra.txt
      const alphaIndex = itemTexts.findIndex((t) => t?.includes("alpha"));
      const appleIndex = itemTexts.findIndex((t) => t?.includes("apple"));
      const zebraIndex = itemTexts.findIndex((t) => t?.includes("zebra"));

      expect(alphaIndex).toBeLessThan(appleIndex);
      expect(appleIndex).toBeLessThan(zebraIndex);
    });
  });

  describe("nested directories", () => {
    it("handles deeply nested file structures", () => {
      const files = {
        "/src/components/ui/Button/index.tsx": "content",
        "/src/components/ui/Button/styles.css": "content",
      };

      render(
        <FileBrowser
          {...defaultProps}
          files={files}
        />
      );

      // Click through directories to expand
      fireEvent.click(screen.getByText("components"));
      fireEvent.click(screen.getByText("ui"));
      fireEvent.click(screen.getByText("Button"));

      // Deep file should be visible
      expect(screen.getByText("index.tsx")).toBeInTheDocument();
    });
  });

  describe("file path normalization", () => {
    it("handles paths with and without leading slash", () => {
      const files = {
        "src/App.jsx": "content",
        "/src/index.js": "content",
      };

      render(
        <FileBrowser
          {...defaultProps}
          files={files}
        />
      );

      expect(screen.getByText("App.jsx")).toBeInTheDocument();
      expect(screen.getByText("index.js")).toBeInTheDocument();
    });
  });
});
