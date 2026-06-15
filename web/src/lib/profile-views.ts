import { api } from './api';

export interface ProfileView {
  id: string;
  profileId: string;
  companyId: string;
  viewedAt: string;
}

export async function getMyProfileViews(): Promise<ProfileView[]> {
  return api<ProfileView[]>('/profiles/me/views');
}
