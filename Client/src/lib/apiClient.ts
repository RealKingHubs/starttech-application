import axios from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://dev-backend-alb-188828417.us-east-1.elb.amazonaws.com";

export const apiClient = axios.create({
    baseURL: API_BASE_URL,
    withCredentials: true, // Crucial for httpOnly cookies
});
