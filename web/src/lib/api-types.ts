import type { components } from './generated/api';

type Schemas = components['schemas'];

// Auth
export type LoginRequest = Schemas['LoginRequest'];
export type LoginResponse = Schemas['LoginResponse'];
export type RegisterRequest = Schemas['RegisterRequest'];
export type RefreshResponse = Schemas['RefreshResponse'];
export type User = Schemas['UserResponse'];

// Reference data
export type Industry = Schemas['IndustryResponse'];
export type Tag = Schemas['TagResponse'];

// Jobs
export type Job = Schemas['JobResponse'];
export type CreateJobRequest = Schemas['CreateJobRequest'];
export type UpdateJobRequest = Schemas['UpdateJobRequest'];

// Profiles & passport
export type Profile = Schemas['ProfileResponse'];
export type PublicProfile = Schemas['PublicProfileResponse'];
export type UpdateProfileResponse = Schemas['UpdateProfileResponse'];
export type Experience = Schemas['Experience'];

// Matching / search / evaluation / application
export type CandidateResult = Schemas['CandidateResult'];
export type CandidateMatch = Schemas['CandidateMatch'];
export type JobMatch = Schemas['JobMatch'];
export type EvaluationResponse = Schemas['EvaluationResponse'];
export type ApplicationResult = Schemas['ApplicationResult'];

// Generic
export type MessageResponse = Schemas['MessageResponse'];
export type VerificationStatusResponse = Schemas['VerificationStatusResponse'];
