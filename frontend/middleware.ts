import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export function middleware(request: NextRequest) {
    if (request.nextUrl.pathname.startsWith('/api/')) {
        // Determine the API backend URL from env or default
        const apiUrl = process.env.API_URL || 'http://localhost:8080';

        // Construct target URL
        // request.nextUrl.pathname includes /api/..., so just append it to source?
        // Wait, API_URL is "http://orchestrator:8080".
        // If we simply concat, we get http://orchestrator:8080/api/pods

        // Note: URL constructor requires absolute URL if base is provided
        const targetUrl = new URL(request.nextUrl.pathname, apiUrl);
        targetUrl.search = request.nextUrl.search;

        // console.log(`[Middleware] Rewriting ${request.url} to ${targetUrl.toString()}`);

        return NextResponse.rewrite(targetUrl);
    }
}

export const config = {
    matcher: '/api/:path*',
}
