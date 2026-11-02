export interface User {
  id: number;
  username: string;
  admin: boolean;
  avatar_id: number | null;
  student_no: string | null;
  real_name: string;
  qq: string | null;
  qq_comment_notifications: boolean;
  category_ids: string;
  created_at: string;
}

export interface Category {
  id: number;
  name: string;
  parent_id: number | null;
  sort: number;
}

export interface Assignment {
  review_status: string;
  id: number;
  name: string;
  category_id: number | null;
  status: string;
  deadline: string | null;
  sort: number;
  description: string;
  summary: string;
  allow_attachments: boolean;
  extensions: string;
  max_size: number;
  multiple: boolean;
  withdraw: boolean;
  require_real_name: boolean;
  require_qq: boolean;
  created_at: string;
  submitters: number;
  submitted: boolean;
}

export interface Attachment {
  id: number;
  user_id: number;
  name: string;
  sha256: string;
  size: number;
  status: string;
  chunk_size: number;
  total_chunks: number;
  received: number[];
  created_at: string;
  usage?: string[];
}

export interface Submission {
  review_status: string;
  assignment_name?: string;
  id: number;
  assignment_id: number;
  user_id: number;
  username: string;
  content: string;
  withdrawn: boolean;
  created_at: string;
  attachments: Attachment[];
}

export interface Page<T> {
  items: T[];
  total: number;
}

export type Settings = Record<string, string>;

export interface ReviewMessage {
  last_notified_at: string | null;
  notification_pending: boolean;
  notification_status: string;
  id: number;
  submission_id: number;
  user_id: number;
  username: string;
  avatar_id: number | null;
  admin: boolean;
  content: string;
  created_at: string;
  attachments: Attachment[];
}
