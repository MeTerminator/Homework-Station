import { StrictMode, useState } from "react";
import { createRoot } from "react-dom/client";
import { SessionProvider } from "../../../src/app/providers/session-provider";
import { ThemeProvider } from "../../../src/app/providers/theme-provider";
import { MarkdownEditor } from "../../../src/components/markdown-editor";
import { Confirm } from "../../../src/components/feedback/confirm";
import { AttachmentPreview } from "../../../src/features/attachments/components/attachment-preview";
import { SubmissionReview } from "../../../src/features/assignments/components/submission-review";
import "../../../src/styles/index.css";

const attachment = { id: 1, name: "preview.png", size: 10 };
const submission = {
  id: 1,
  username: "test",
  created_at: "2026-11-02T00:00:00Z",
  content: "Original submission",
  attachments: [attachment],
  review_status: "pending",
  withdrawn: false,
};

function Fixture() {
  const [review, setReview] = useState(false);
  const [confirm, setConfirm] = useState(false);
  const [editor, setEditor] = useState(false);
  const [value, setValue] = useState("");
  return (
    <main style={{ minHeight: "300vh", padding: 24 }}>
      <button onClick={() => setReview(true)}>Open review</button>
      <button onClick={() => setConfirm(true)}>Open confirm</button>
      <button onClick={() => setEditor((v) => !v)}>Toggle editor</button>
      <AttachmentPreview file={attachment} />
      {editor && <MarkdownEditor value={value} onChange={setValue} label="Editor" />}
      {review && (
        <SubmissionReview
          submission={submission}
          onClose={() => setReview(false)}
          onChange={() => {}}
        />
      )}
      <Confirm
        open={confirm}
        onClose={() => setConfirm(false)}
        title="Confirm test"
        description="Scroll lock regression"
        onConfirm={async () => {}}
      />
    </main>
  );
}

createRoot(document.getElementById("root")).render(
  <StrictMode>
    <ThemeProvider>
      <SessionProvider>
        <Fixture />
      </SessionProvider>
    </ThemeProvider>
  </StrictMode>,
);
