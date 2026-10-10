# Issue #009 - Goal-Based Product Search on E-Shop

**Type**: enhancement
**Priority**: medium
**Status**: ✅ Implemented
**GitHub**: #9
**Branch**: `feature/010-mcp-http-transport`
**Depends on**: Issue #008
**Related UC**: [UC-001](../../use-cases/UC-001-goal-based-product-search.md)

## Feature Description

Implementation of use case UC-001 — Goal-Based Product Search on E-Shop.

Enable an AI assistant (e.g., Claude Code) to accept a natural language product goal from the user, navigate an e-shop (initially Alza.cz) in a visible browser, search for matching products, and present the result (product name, price, link) while the user observes the automation in real time.

The flow leverages ptx-agent browser MCP tools (from Issue #008): launch browser in headed mode, navigate to e-shop, search, sort/filter results, extract product data, and leave the browser open on the product detail page.

## Acceptance Criteria

1. [ ] User can describe a product need in natural language and get a matching product from Alza.cz
2. [ ] AI assistant navigates Alza.cz, searches, and extracts product data using ptx-agent browser MCP tools
3. [ ] Product name, price, and URL are returned to the user
4. [ ] Product detail page is displayed in a visible browser window so the user can see the result
5. [ ] User can observe the entire automation process in real time (headed mode)

## Complexity Estimate

Medium — the browser control primitives exist (Issue #008), but reliable e-shop navigation (search, filtering, data extraction, cookie/consent handling) requires robust element interaction and error handling.
