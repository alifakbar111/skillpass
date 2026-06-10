export interface Company {
  id: string;
  companyName: string;
  website?: string;
  industry: string;
  description?: string;
  verificationDocs?: Record<string, string>;
  createdAt: string;
}
