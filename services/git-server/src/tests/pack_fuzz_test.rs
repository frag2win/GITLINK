#[cfg(test)]
mod tests {
    use git_server::pack::PackParser;

    #[test]
    fn fuzz_pack_parser_corrupt_header() {
        // Simulating a fuzzing test checking resilience against corrupt headers
        let corrupt_data = vec![0x50, 0x41, 0x43, 0x4B, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00];
        let mut parser = PackParser::new(&corrupt_data[..]);
        let result = parser.parse();
        assert!(result.is_err(), "Parser should fail on corrupt header without panicking");
    }

    #[test]
    fn fuzz_pack_parser_truncated_body() {
        let truncated_data = vec![0x50, 0x41, 0x43, 0x4B, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x01, 0x90];
        let mut parser = PackParser::new(&truncated_data[..]);
        let result = parser.parse();
        assert!(result.is_err(), "Parser should gracefully reject truncated body");
    }

    #[test]
    fn fuzz_pack_parser_invalid_checksum() {
        let invalid_checksum_data = vec![0x50, 0x41, 0x43, 0x4B, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD];
        let mut parser = PackParser::new(&invalid_checksum_data[..]);
        let result = parser.parse();
        assert!(result.is_err(), "Parser should reject packs with invalid SHA-1 checksums");
    }
}
