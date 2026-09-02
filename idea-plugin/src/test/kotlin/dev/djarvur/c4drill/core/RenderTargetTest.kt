package dev.djarvur.c4drill.core

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

/**
 * Port of the #27 extension's renderTarget tests. The expected virtual
 * layout mirrors internal/output/writer.go Write and the link computation in
 * internal/graph/path.go ComputeExploreURL, verified against real CLI output
 * for examples/cloud-system:
 *
 *	cloud-system.svg                 (C1)
 *	cloud-system/cloud.svg           (C2 "cloud")
 *	cloud-system/amazon/rds.svg      (C3 "amazon.rds")
 *
 * with hrefs like "cloud-system/amazon/rds.svg" from C1, "amazon.svg" and
 * "../cloud-system.svg" from C2, "../amazon.svg" from C3.
 */
class RenderTargetTest {
    @Test
    fun `virtual file path mirrors the cli output layout`() {
        assertEquals("cloud-system.svg", virtualFilePath("", "cloud-system"))
        assertEquals("cloud-system/cloud.svg", virtualFilePath("cloud", "cloud-system"))
        assertEquals("cloud-system/amazon/rds.svg", virtualFilePath("amazon.rds", "cloud-system"))
        assertEquals("m/a/b/c.svg", virtualFilePath("a.b.c", "m"))
    }

    @Test
    fun `basename strips scheme, directories and the last extension only`() {
        assertEquals("cloud-system", basenameOfUri("file:///w/models/cloud-system.toml"))
        assertEquals("diagram", basenameOfUri("file:///w/diagram.c4d"))
        assertEquals("my model.architecture", basenameOfUri("file:///w/my%20model.architecture.toml"))
        assertEquals("noext", basenameOfUri("file:///a/noext"))
    }

    @Test
    fun `c1 drill-down links resolve to full dotted targets`() {
        assertEquals("amazon.rds", resolveRenderTarget("", "cloud-system", "cloud-system/amazon/rds.svg"))
        assertEquals("cloud", resolveRenderTarget("", "cloud-system", "cloud-system/cloud.svg"))
    }

    @Test
    fun `c2-c3 sibling, descendant and back links resolve`() {
        // From C2 "cloud" (file cloud-system/cloud.svg).
        assertEquals("amazon", resolveRenderTarget("cloud", "cloud-system", "amazon.svg"))
        assertEquals("", resolveRenderTarget("cloud", "cloud-system", "../cloud-system.svg"))

        // From C3 "amazon.rds" (file cloud-system/amazon/rds.svg).
        assertEquals("amazon", resolveRenderTarget("amazon.rds", "cloud-system", "../amazon.svg"))
        assertEquals("cloud", resolveRenderTarget("amazon.rds", "cloud-system", "../cloud.svg"))
        assertEquals("", resolveRenderTarget("amazon.rds", "cloud-system", "../../cloud-system.svg"))
    }

    @Test
    fun `url-encoded segments are decoded into the dotted path`() {
        assertEquals("a b.c~d", resolveRenderTarget("", "m", "m/a%20b/c~d.svg"))
    }

    @Test
    fun `query and fragment suffixes are tolerated`() {
        assertEquals("a.b", resolveRenderTarget("", "m", "m/a/b.svg#frag"))
        assertEquals("a.b", resolveRenderTarget("", "m", "m/a/b.svg?x=1"))
    }

    @Test
    fun `external http-s reference links are not internal targets`() {
        assertNull(resolveRenderTarget("", "m", "https://example.com/docs"))
        assertNull(resolveRenderTarget("a", "m", "http://example.com/x.svg"))
        assertNull(resolveRenderTarget("", "m", "mailto:someone@example.com"))
    }

    @Test
    fun `links escaping the diagram tree resolve to null`() {
        assertNull(resolveRenderTarget("a", "m", "../../../../etc/passwd.svg"))
        assertNull(resolveRenderTarget("", "m", ""))
        assertNull(resolveRenderTarget("a", "m", ""))
    }

    @Test
    fun `non-svg targets inside the tree are rejected`() {
        assertNull(resolveRenderTarget("", "cloud-system", "cloud-system/cloud.png"))
        assertNull(resolveRenderTarget("", "cloud-system", "unrelated/cloud.svg"))
    }
}
