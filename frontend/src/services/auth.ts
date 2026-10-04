import type {
  LoginRequest,
  LoginResponse,
  RefreshTokenResponse,
  RegisterRequest,
  User,
  UserURLItem,
} from "../types/auth";
import request from "./apiClient";

export async function register(req: RegisterRequest): Promise<User> {
  return request<User>("/api/v1/auth/register", {
    method: "POST",
    body: JSON.stringify(req),
  });
}

export async function login(req: LoginRequest): Promise<LoginResponse> {
  return request<LoginResponse>("/api/v1/auth/login", {
    method: "POST",
    body: JSON.stringify(req),
  });
}

export async function refreshToken(): Promise<RefreshTokenResponse> {
  return request<RefreshTokenResponse>("/api/v1/auth/refresh", {
    method: "POST",
  });
}

export async function logout(): Promise<void> {
  return request<void>("/api/v1/auth/logout", {
    method: "POST",
  });
}

export async function getUserURLs(accessToken: string): Promise<UserURLItem[]> {
  return request<UserURLItem[]>("/api/v1/urls", {
    method: "GET",
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
}
