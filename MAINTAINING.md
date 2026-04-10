# Maintenance instructions and checklist for maintainers. Forks are welcome just don't come with knives


## 1. Setting up your Fork 🍴
If you've forked this repository, follow these steps to keep it in sync and prepare for your own changes:

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/easyp2p.git
cd easyp2p

# Add the original repository as 'upstream' to stay in sync
git remote add upstream https://github.com/opensource-jeff/easyp2p.git

# Create a branch for your features
git checkout -b my-new-feature
```

## 2. Dependency Management 📦
`libp2p` moves fast. You should periodically check for updates to keep your fork secure and performant.

- **Update Dependencies**: Run `go get -u ./...` followed by `go mod tidy`.
- **Breaking Changes**: `libp2p` occasionally introduces breaking changes. Always test the `examples/` after updating dependencies to ensure they still work.

## 3. Testing Strategy 🧪
To maintain "easiness," ensure that new changes don't break the existing API.

- **Run Tests**: Use `go test ./...` regularly.
- **Add Integration Tests**: When adding a new feature (e.g., a custom transport), add a small test in `_test.go` or an example in `examples/` that proves it works with two nodes.

## 4. Keeping it "Easy" 🧘
The "Easy" in `easyp2p` is its core philosophy.
- **Avoid "Feature Creep"**: If a new feature requires complex networking knowledge (like DHT routing tables), wrap it in a simpler function or provide sensible defaults.
- **Documentation First**: Update `README.md` and add a simple example in `examples/` for every new capability.

## 5. Releasing Versions & Module Names 🏷️
When you reach a stable point, tag a version so others can use your fork.

```bash
git tag v1.0.0
git push origin v1.0.0
```

**Note on Go Modules**: If you want others to import your fork as a library, you should update the module name in `go.mod` to `github.com/YOUR_USERNAME/easyp2p`. This allows users to `go get` your specific version.

## 6. Handling Contributions 🤝
If you want to accept contributions to your fork:
- **Issues**: Use GitHub Issues to track bugs or feature requests.
- **Pull Requests**: Review code to ensure it follows the "simple" philosophy before merging.
- **Contributing Back**: If you've fixed a bug or added a great feature, consider opening a PR back to the upstream repository!
