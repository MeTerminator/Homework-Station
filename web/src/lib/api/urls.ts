export const fileURL = (id: number, inline = false) =>
  `/api/attachments/${id}/download${inline ? "?inline=true" : ""}`;
