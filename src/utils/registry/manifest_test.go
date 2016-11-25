package registry

import (
	"testing"

	"github.com/docker/distribution/manifest/schema2"
)

func TestUnMarshal(t *testing.T) {
	b := []byte(`{  
   "schemaVersion":2,
   "mediaType":"application/vnd.docker.distribution.manifest.v2+json",
   "config":{  
      "mediaType":"application/vnd.docker.container.image.v1+json",
      "size":1473,
      "digest":"sha256:c54a2cc56cbb2f04003c1cd4507e118af7c0d340fe7e2720f70976c4b75237dc"
   },
   "layers":[  
      {  
         "mediaType":"application/vnd.docker.image.rootfs.diff.tar.gzip",
         "size":974,
         "digest":"sha256:c04b14da8d1441880ed3fe6106fb2cc6fa1c9661846ac0266b8a5ec8edf37b7c"
      }
   ]
}`)

	manifest, _, err := UnMarshal(schema2.MediaTypeManifest, b)
	if err != nil {
		t.Fatalf("failed to parse manifest: %v", err)
	}

	refs := manifest.References()
	if len(refs) != 1 {
		t.Fatalf("unexpected length of reference: %d != %d", len(refs), 1)
	}

	digest := "sha256:c04b14da8d1441880ed3fe6106fb2cc6fa1c9661846ac0266b8a5ec8edf37b7c"
	if refs[0].Digest.String() != digest {
		t.Errorf("unexpected digest: %s != %s", refs[0].Digest.String(), digest)
	}
}
