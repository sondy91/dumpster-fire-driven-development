"""API client for metrics-api integration."""

import asyncio
import hashlib
import os
from datetime import datetime
from typing import Any

import httpx

# Simple in-memory cache for metrics data
_METRICS_CACHE = {}
_CACHE_TTL_SECONDS = 3600  # 1 hour


def clear_metrics_cache():
    """Clear the entire metrics cache."""
    global _METRICS_CACHE
    cache_size = len(_METRICS_CACHE)
    _METRICS_CACHE.clear()
    return cache_size


class MetricsAPIClient:
    """Client for interacting with metrics-api."""

    def __init__(self, base_url: str | None = None, timeout: int = 30):
        """Initialize API client.

        Args:
            base_url: Base URL for metrics-api (default: env var or localhost)
            timeout: Request timeout in seconds
        """
        self.base_url = base_url or os.getenv("METRICS_API_URL", "http://localhost:8080")
        self.timeout = timeout

        transport = httpx.HTTPTransport(retries=3)
        self.client = httpx.Client(
            base_url=self.base_url,
            timeout=timeout,
            transport=transport,
        )
        self.async_client = httpx.AsyncClient(
            base_url=self.base_url,
            timeout=timeout,
            transport=httpx.AsyncHTTPTransport(retries=3),
        )

    def health(self) -> dict[str, Any]:
        """Check API health status.

        Returns:
            Health check response

        Raises:
            httpx.HTTPError: If API is unreachable
        """
        response = self.client.get("/api/v1/health")
        response.raise_for_status()
        return response.json()

    def get_jira_metrics(self, start_date: str, end_date: str) -> dict[str, Any]:
        """Get Jira metrics for date range.

        Args:
            start_date: Start date (YYYY-MM-DD)
            end_date: End date (YYYY-MM-DD)

        Returns:
            Jira metrics response

        Raises:
            httpx.HTTPError: If API call fails
        """
        response = self.client.get(
            "/api/v1/metrics/jira",
            params={"start_date": start_date, "end_date": end_date},
        )
        response.raise_for_status()
        return response.json()

    def get_github_metrics(self, start_date: str, end_date: str) -> dict[str, Any]:
        """Get GitHub metrics for date range.

        Args:
            start_date: Start date (YYYY-MM-DD)
            end_date: End date (YYYY-MM-DD)

        Returns:
            GitHub metrics response

        Raises:
            httpx.HTTPError: If API call fails
        """
        response = self.client.get(
            "/api/v1/metrics/github",
            params={"start_date": start_date, "end_date": end_date},
        )
        response.raise_for_status()
        return response.json()

    def get_git_metrics(
        self,
        start_date: str,
        end_date: str,
        e_number: str | None = None,
    ) -> dict[str, Any]:
        """Get Git metrics for date range.

        Args:
            start_date: Start date (YYYY-MM-DD)
            end_date: End date (YYYY-MM-DD)
            e_number: Optional E-number to query specific user

        Returns:
            Git metrics response

        Raises:
            httpx.HTTPError: If API call fails
        """
        params = {"start_date": start_date, "end_date": end_date}
        if e_number:
            params["e_number"] = e_number

        response = self.client.get("/api/v1/metrics/git", params=params)
        response.raise_for_status()
        return response.json()

    def is_available(self) -> bool:
        """Check if API is available.

        Returns:
            True if API is reachable and healthy, False otherwise
        """
        try:
            self.health()
            return True
        except httpx.HTTPError:
            return False

    async def get_all_metrics_parallel(
        self, start_date: str, end_date: str, e_number: str | None = None
    ) -> dict[str, Any]:
        """Fetch all metrics in parallel for faster loading.

        Server-side caching: Results cached for 1 hour based on date range.

        Args:
            start_date: Start date (YYYY-MM-DD)
            end_date: End date (YYYY-MM-DD)
            e_number: Optional E-number for git metrics

        Returns:
            Dict with 'jira', 'github', 'git' keys, each containing result or error
        """
        # Check cache first
        cache_key = hashlib.md5(f"{start_date}:{end_date}:{e_number}".encode()).hexdigest()

        if cache_key in _METRICS_CACHE:
            cached_data, cached_time = _METRICS_CACHE[cache_key]
            age_seconds = (datetime.now() - cached_time).total_seconds()
            if age_seconds < _CACHE_TTL_SECONDS:
                print(f"✅ Metrics cache hit! Age: {age_seconds:.0f}s / {_CACHE_TTL_SECONDS}s")
                return cached_data
            else:
                print(f"⏰ Metrics cache expired (age: {age_seconds:.0f}s)")
                del _METRICS_CACHE[cache_key]

        async def fetch_jira():
            try:
                response = await self.async_client.get(
                    "/api/v1/metrics/jira",
                    params={"start_date": start_date, "end_date": end_date},
                )
                response.raise_for_status()
                return response.json()
            except Exception as e:
                return {"error": str(e)}

        async def fetch_github():
            try:
                response = await self.async_client.get(
                    "/api/v1/metrics/github",
                    params={"start_date": start_date, "end_date": end_date},
                )
                response.raise_for_status()
                return response.json()
            except Exception as e:
                return {"error": str(e)}

        async def fetch_git():
            try:
                params = {"start_date": start_date, "end_date": end_date}
                if e_number:
                    params["e_number"] = e_number
                response = await self.async_client.get("/api/v1/metrics/git", params=params)
                response.raise_for_status()
                return response.json()
            except Exception as e:
                return {"error": str(e)}

        jira_result, github_result, git_result = await asyncio.gather(fetch_jira(), fetch_github(), fetch_git())

        result = {
            "jira": jira_result if "error" not in jira_result else None,
            "github": github_result if "error" not in github_result else None,
            "git": git_result if "error" not in git_result else None,
            "errors": [f"Jira: {jira_result['error']}" for _ in [""] if "error" in jira_result]
            + [f"GitHub: {github_result['error']}" for _ in [""] if "error" in github_result]
            + [f"Git: {git_result['error']}" for _ in [""] if "error" in git_result],
        }

        # Cache the result
        _METRICS_CACHE[cache_key] = (result, datetime.now())
        print(f"💾 Cached metrics data (key: {cache_key[:8]}...)")

        return result

    def close(self):
        """Close the HTTP clients."""
        self.client.close()
        try:
            loop = asyncio.get_running_loop()
            # Already in async context, schedule closing
            loop.create_task(self.async_client.aclose())
        except RuntimeError:
            # No running loop, safe to use run_until_complete
            asyncio.get_event_loop().run_until_complete(self.async_client.aclose())

    def __enter__(self):
        """Context manager entry."""
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.close()
        return False
