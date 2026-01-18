import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import "@testing-library/jest-dom";
import CodeEditor from "../CodeEditor";

describe("CodeEditor", () => {
  const defaultProps = {
    filePath: "/src/App.jsx",
    content: 'export default function App() { return <div>Hello</div>; }',
    onChange: vi.fn(),
    onSave: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("rendering", () => {
    it("renders empty state when no file is selected", () => {
      render(
        <CodeEditor
          {...defaultProps}
          filePath={null}
          content=""
        />
      );

      expect(screen.getByText("Select a file to edit")).toBeInTheDocument();
    });

    it("renders file name in header", () => {
      render(<CodeEditor {...defaultProps} />);

      expect(screen.getByText("App.jsx")).toBeInTheDocument();
    });

    it("renders language badge for JavaScript", () => {
      render(<CodeEditor {...defaultProps} />);

      expect(screen.getByText("javascript")).toBeInTheDocument();
    });

    it("renders correct language for TypeScript files", () => {
      render(
        <CodeEditor
          {...defaultProps}
          filePath="/src/App.tsx"
        />
      );

      expect(screen.getByText("typescript")).toBeInTheDocument();
    });

    it("renders correct language for CSS files", () => {
      render(
        <CodeEditor
          {...defaultProps}
          filePath="/src/styles.css"
        />
      );

      expect(screen.getByText("css")).toBeInTheDocument();
    });

    it("renders correct language for JSON files", () => {
      render(
        <CodeEditor
          {...defaultProps}
          filePath="/package.json"
        />
      );

      expect(screen.getByText("json")).toBeInTheDocument();
    });

    it("renders line count", () => {
      const multiLineContent = "line1\nline2\nline3";
      render(
        <CodeEditor
          {...defaultProps}
          content={multiLineContent}
        />
      );

      expect(screen.getByText("3 lines")).toBeInTheDocument();
    });

    it("renders file path in status bar", () => {
      render(<CodeEditor {...defaultProps} />);

      expect(screen.getByText("/src/App.jsx")).toBeInTheDocument();
    });

    it("renders save shortcut hint when editable", () => {
      render(<CodeEditor {...defaultProps} />);

      expect(screen.getByText("Cmd+S to save")).toBeInTheDocument();
    });

    it("does not render save shortcut when readOnly", () => {
      render(<CodeEditor {...defaultProps} readOnly />);

      expect(screen.queryByText("Cmd+S to save")).not.toBeInTheDocument();
    });
  });

  describe("editing", () => {
    it("allows editing when not readOnly", () => {
      render(<CodeEditor {...defaultProps} />);

      const textarea = screen.getByRole("textbox");
      expect(textarea).not.toHaveAttribute("readonly");
    });

    it("prevents editing when readOnly", () => {
      render(<CodeEditor {...defaultProps} readOnly />);

      const textarea = screen.getByRole("textbox");
      expect(textarea).toHaveAttribute("readonly");
    });

    it("shows dirty indicator when content changes", async () => {
      render(<CodeEditor {...defaultProps} />);

      const textarea = screen.getByRole("textbox");
      fireEvent.change(textarea, { target: { value: "new content" } });

      // Dirty indicator should appear
      await waitFor(() => {
        expect(screen.getByTitle("Unsaved changes")).toBeInTheDocument();
      });
    });

    it("shows save button when dirty", async () => {
      render(<CodeEditor {...defaultProps} />);

      const textarea = screen.getByRole("textbox");
      fireEvent.change(textarea, { target: { value: "modified content" } });

      await waitFor(() => {
        expect(screen.getByText("Save")).toBeInTheDocument();
      });
    });

    it("calls onChange and onSave when save button clicked", async () => {
      const onChange = vi.fn();
      const onSave = vi.fn();

      render(
        <CodeEditor
          {...defaultProps}
          onChange={onChange}
          onSave={onSave}
        />
      );

      const textarea = screen.getByRole("textbox");
      fireEvent.change(textarea, { target: { value: "new content" } });

      const saveButton = await screen.findByText("Save");
      fireEvent.click(saveButton);

      expect(onChange).toHaveBeenCalledWith("new content");
      expect(onSave).toHaveBeenCalled();
    });
  });

  describe("keyboard shortcuts", () => {
    it("inserts spaces when Tab is pressed", async () => {
      render(<CodeEditor {...defaultProps} content="" />);

      const textarea = screen.getByRole("textbox");
      fireEvent.keyDown(textarea, { key: "Tab" });

      await waitFor(() => {
        expect((textarea as HTMLTextAreaElement).value).toBe("  ");
      });
    });

    it("handles Cmd+S to save", async () => {
      const onChange = vi.fn();
      const onSave = vi.fn();

      render(
        <CodeEditor
          {...defaultProps}
          onChange={onChange}
          onSave={onSave}
        />
      );

      const textarea = screen.getByRole("textbox");
      fireEvent.change(textarea, { target: { value: "new content" } });

      // Simulate Cmd+S
      fireEvent.keyDown(window, { key: "s", metaKey: true });

      await waitFor(() => {
        expect(onChange).toHaveBeenCalled();
        expect(onSave).toHaveBeenCalled();
      });
    });
  });

  describe("line numbers", () => {
    it("renders correct number of line numbers", () => {
      const content = "line1\nline2\nline3\nline4\nline5";
      render(<CodeEditor {...defaultProps} content={content} />);

      // Check line numbers are rendered
      expect(screen.getByText("1")).toBeInTheDocument();
      expect(screen.getByText("5")).toBeInTheDocument();
    });
  });

  describe("language detection", () => {
    const testCases = [
      { file: "app.js", expected: "javascript" },
      { file: "app.jsx", expected: "javascript" },
      { file: "app.ts", expected: "typescript" },
      { file: "app.tsx", expected: "typescript" },
      { file: "styles.css", expected: "css" },
      { file: "index.html", expected: "html" },
      { file: "data.json", expected: "json" },
      { file: "readme.md", expected: "markdown" },
      { file: "unknown.xyz", expected: "plaintext" },
    ];

    testCases.forEach(({ file, expected }) => {
      it(`detects ${expected} for ${file}`, () => {
        render(
          <CodeEditor
            {...defaultProps}
            filePath={`/src/${file}`}
          />
        );

        expect(screen.getByText(expected)).toBeInTheDocument();
      });
    });
  });
});
