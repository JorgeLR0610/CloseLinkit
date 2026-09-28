export interface User {
  id: string;
  email: string;
  email_verified_at?: string | null;
  created_at: string;
  updated_at?: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
}

export type RegisterResponse = User;

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token?: string;
  expires_in: number;
  user: User;
}

export interface RefreshTokenResponse {
  access_token: string;
  refresh_token?: string;
  expires_in: number;
  user: User;
}

export interface UserURLItem {
  original_url: string;
  short_code: string;
  short_url: string;
  created_at: string;
  click_count: number;
}
