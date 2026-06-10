export interface Job {
  id: string;
  title: string;
  industry: string;
  location?: string;
  experienceLevel?: string;
  status: string;
  createdAt: string;
}

export interface Industry {
  ID: string;
  Name: string;
  Description: string | null;
}
