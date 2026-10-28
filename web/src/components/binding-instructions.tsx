export function BindingInstructions({ text }: { text: string }) {
  return (
    <p className="muted small binding-instructions">
      {text.split(/(`[^`\n]+`)/g).map((part, index) =>
        part.startsWith("`") && part.endsWith("`") && part.length > 2 ? (
          <code className="binding-command" key={index}>
            {part.slice(1, -1)}
          </code>
        ) : (
          part
        ),
      )}
    </p>
  );
}
