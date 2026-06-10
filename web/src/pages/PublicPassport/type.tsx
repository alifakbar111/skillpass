export interface PassportData {
  name: string;
  avatarUrl?: string;
  headline?: string;
  about?: string;
  yearsOfExperience?: number;
  viewCount?: number;
  experiences: Array<{
    type: string;
    title: string;
    organization: string;
    startDate: string;
    endDate?: string;
    isCurrent: boolean;
    description?: string;
    skillsUsed?: string[];
    url?: string;
  }>;
}
