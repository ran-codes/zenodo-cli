Zenodo MCP server — provides read-only access to the Zenodo research repository.

## Search tips

- records_search uses Elasticsearch query syntax.
- To find records by ORCID, always search BOTH creators and contributors:
  q: "creators.orcid:XXXX-XXXX-XXXX-XXXX OR contributors.orcid:XXXX-XXXX-XXXX-XXXX"
  A person can be listed as a creator on some records and a contributor on others.
- To find records by name: q: "creators.name:\"LastName, FirstName\""
- To filter by type: q: "resource_type.type:dataset"
- To filter by community: use the community parameter instead of the query.
- records_list returns only the authenticated user's uploads/drafts (deposit API). For a complete view of records associated with a person, use records_search with their ORCID.
