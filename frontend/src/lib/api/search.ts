import { SearchNodes, SearchProperties, SearchByPath, GetNodesByDepth, GetIndexHealth } from '../../../wailsjs/go/api/SearchService';

// Types matching Go structs
export interface SearchRequest {
    query: string;
    limit: number;
}

export interface SearchResult {
    nodeId: string;
    name: string;
    description: string;
    path: string;
    depth: number;
    childCount: number;
    rank: number;
    snippet?: string;
}

export interface PropertySearchResult {
    nodeId: string;
    key: string;
    value: string;
    rank: number;
}

export interface SearchResponse {
    results: SearchResult[];
    count: number;
}

export interface PropertySearchResponse {
    results: PropertySearchResult[];
    count: number;
}

// Search API wrapper
export class SearchAPI {
    /**
     * Search nodes by name, description, or path content
     */
    static async searchNodes(query: string, limit: number = 50): Promise<SearchResponse> {
        const request: SearchRequest = { query, limit };
        return await SearchNodes(request);
    }

    /**
     * Search node properties by key or value
     */
    static async searchProperties(query: string, limit: number = 50): Promise<PropertySearchResponse> {
        const request: SearchRequest = { query, limit };
        return await SearchProperties(request);
    }

    /**
     * Search nodes by path prefix
     */
    static async searchByPath(pathPrefix: string, limit: number = 50): Promise<SearchResponse> {
        return await SearchByPath(pathPrefix, limit);
    }

    /**
     * Get all nodes at a specific depth level
     */
    static async getNodesByDepth(depth: number, limit: number = 50): Promise<SearchResponse> {
        return await GetNodesByDepth(depth, limit);
    }

    /**
     * Get search index health status
     */
    static async getIndexHealth(): Promise<Record<string, any>> {
        return await GetIndexHealth();
    }

    /**
     * Perform a combined search across nodes and properties
     */
    static async searchAll(query: string, limit: number = 50): Promise<{
        nodes: SearchResponse;
        properties: PropertySearchResponse;
    }> {
        const [nodes, properties] = await Promise.all([
            this.searchNodes(query, limit),
            this.searchProperties(query, limit)
        ]);

        return { nodes, properties };
    }

    /**
     * Search with auto-completion suggestions
     */
    static async searchWithSuggestions(query: string, limit: number = 10): Promise<SearchResponse> {
        // For auto-complete, use shorter limit and add wildcard if not already present
        const searchQuery = query.endsWith('*') ? query : query + '*';
        return await this.searchNodes(searchQuery, limit);
    }
}

// Legacy functions for backwards compatibility
export async function search(q: string) {
    return SearchAPI.searchNodes(q);
}

export async function rebuildIndex() {
    // TODO: Implement rebuild functionality when available
    throw new Error('Index rebuild not yet implemented');
}
