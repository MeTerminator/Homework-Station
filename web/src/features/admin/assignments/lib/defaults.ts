import type { Assignment } from "@/types/models";
export const blankAssignment: Assignment = {
  id: 0,
  name: "",
  category_id: null,
  status: "hidden",
  deadline: null,
  sort: 0,
  description: "",
  allow_attachments: true,
  extensions: "",
  max_size: 20 * 1024 ** 2,
  multiple: true,
  withdraw: true,
  created_at: "",
  submitters: 0,
  submitted: false,
};
