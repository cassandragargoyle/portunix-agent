# UC-001 — Goal-Based Product Search on E-Shop

**Status**: ✅ Covered by Issue
**Related Issue**: [#009](../issues/internal/done/009-goal-based-product-search.md)
**Created**: 2026-02-21

## Actor

User interacting with Claude Code (or another AI assistant) that has access to ptx-agent MCP tools.

## Preconditions

- ptx-agent is running with MCP server enabled
- Browser module is available (Playwright browsers installed)
- AI assistant has access to ptx-agent MCP tools (browser control)

## Main Flow

1. User describes a product goal in natural language (e.g., "Find the cheapest LG washing machine on Alza and show me the price")
2. AI assistant interprets the goal and determines target e-shop (Alza.cz) and search criteria
3. AI assistant launches browser in headed (visible) mode via `agent_browser_launch` with `headless: false` so the user can observe the automation in real time
4. AI assistant navigates to the e-shop via `agent_browser_goto`
5. AI assistant locates the search field using `agent_browser_elements` and types the product query via `agent_browser_type`
6. AI assistant submits the search and waits for results
7. AI assistant reads the search results using `agent_browser_content` and/or `agent_browser_elements`
8. AI assistant applies sorting/filtering (e.g., sort by price) if needed by interacting with page controls
9. AI assistant extracts product information (name, price, availability, URL) from the results
10. AI assistant navigates to the product detail page via `agent_browser_click`
11. AI assistant reads product details from the visible page
12. AI assistant presents the result to the user: product name, price, and link — the browser remains open on the product page for the user to inspect

## Alternative Flows

- Product not found: AI assistant informs user that no matching products were found and suggests alternative search terms
- Multiple matching products: AI assistant presents a short list of top matches with prices for user to choose from
- E-shop blocks automation: AI assistant reports the issue and suggests manual browsing
- Cookie/consent dialog appears: AI assistant dismisses the dialog before proceeding with search

## Postconditions

- User receives product information (name, price, availability) matching their goal
- Browser remains open on the product detail page, visible to the user

## Acceptance Criteria

1. [ ] User can describe a product need in natural language and get a matching product from Alza.cz
2. [ ] AI assistant navigates Alza.cz, searches, and extracts product data using ptx-agent browser MCP tools
3. [ ] Product name, price, and URL are returned to the user
4. [ ] Product detail page is displayed in a visible browser window so the user can see the result
5. [ ] User can observe the entire automation process in real time (headed mode)

## Notes

- This UC validates the end-to-end browser control workflow from Issue #008
- The AI assistant (Claude Code) orchestrates the flow; ptx-agent provides the browser control primitives via MCP
- Initial target e-shop is Alza.cz but the pattern is generalizable to other e-shops
