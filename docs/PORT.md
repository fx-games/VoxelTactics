# Port to VoxelLib

Branch `voxellib-port`. Each step ends with `go build ./... && go vet ./... && go test ./...`
passing for the whole module, then a commit. Exception: the `client` package is
rebuilt in steps 5-7 and does not build in between.

## Decisions
- Desktop only; the WASM/WebGL build is dropped.
- Blocks are VoxelLib block states (names, properties, shapes). Never call them
  anything but "blocks" in game text, docs or identifiers.
- Textures: one set, Star Odyssey (`assets/textures/blocks/star_odyssey`), served as a
  resource pack over the base pack's block models.
- Maps are region worlds (VoxelLib `voxel/anvil`) plus the existing `.meta` JSON.
  The server reads them with `anvil` into a `voxel.World`; the client streams the same
  folder through `voxel/engine` with `Bounds` and `ReadOnly`.
- VoxelLib does navigation (`nav.Graph`, `nav.Reach` for the move range), raycasts
  (`physics.Walk`, `physics.RaycastBox`) and mesh hits (`model.RaycastPose`, bones →
  body zones). Game rules (LOS corners, postures, AP, damage zones, occupancy) stay here.

## Steps
- [x] 1. World: `engine/voxel.Map` wraps a VoxelLib `voxel.World` (bounds, occupants);
      chunk storage, greedy mesher, game lighting and GL code go.
- [x] 2. Maps: convert `map.bin`, `test-map.bin` and the `.construction` imports
      (keeping block properties) to region worlds; editor saves regions.
- [ ] 3. Movement: `nav.Reach`/`nav.Graph` replace `engine/path` and
      `GetNeighborsForGroundMovement`; footprint and occupancy via `Passable`.
- [ ] 4. Raycasts: `physics.Walk`/`RaycastBox` replace the DDA; unit hits use
      `model.RaycastPose`. Delete the dead GJK/EPA/AABB solver.
- [ ] 5. Client: window, loop, terrain and sky through `voxel/engine` + `voxel/render`;
      iso and free cameras on `render.Camera`; cutaway height.
- [ ] 6. Units, projectiles, selector: glTF via `importer`/`model` + `render.ModelRenderer`.
- [ ] 7. Overlays: decals and block highlights become a VoxelLib `voxel/render` feature
      (the game only says where); lines, smoke, HUD, menus, fonts on VoxelLib's render/ui;
      audio on `audio`. `client/decals.go` and `client/highlights.go` are the old GL
      versions, parked there until then.
- [ ] 8. Remove `engine/glhf`, `engine/util` GL code, shaders, `main_js.go`, `build_web.sh`,
      `go-gl` deps. Screenshot check of deploy, move, shoot, throw, editor.
