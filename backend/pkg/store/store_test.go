package store

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/safe-waters/retro-simply/backend/pkg/data"
)

func TestRedisDataStoreState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Name          string
		OldState      []byte
		State         []byte
		ExpectedState []byte
	}{
		{
			Name: "Same State",
			OldState: []byte(`
{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": null
}
`,
			),
			State: []byte(`
{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": null
}
`,
			),
			ExpectedState: []byte(`

{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": null
}
`,
			),
		},
		{
			Name: "New Group",
			OldState: []byte(`
{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": null
}
`,
			),
			State: []byte(`
{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                },
                {
                    "id": "some-uuid",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": null
}
`,
			),
			ExpectedState: []byte(`
{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                },
                {
                    "id": "some-uuid",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": null
}
`,
			),
		},
		{
			Name: "New Group With New Card",
			OldState: []byte(`
{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": null
}
`,
			),
			State: []byte(`
{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                },
                {
                    "id": "some-uuid",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "anotheruuid-pk-0",
                            "columnId": "0",
                            "message": "different message",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "some-uuid",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": null
}
`,
			),
			ExpectedState: []byte(`
{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                },
                {
                    "id": "some-uuid",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "anotheruuid-pk-0",
                            "columnId": "0",
                            "message": "different message",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "some-uuid",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": null
}
`,
			),
		},
		{
			Name: "Upvote",
			OldState: []byte(`
{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": null
}
`,
			),
			State: []byte(`
{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 1,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": {
        "title": "upVote",
        "oldCard": {
            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
            "columnId": "0",
            "message": "hello",
            "numVotes": 0,
            "isEditable": false,
            "groupId": "default",
            "isDeleted": false,
            "lastModified": 1617661857846
        },
        "newCard": {
            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
            "columnId": "0",
            "message": "hello",
            "numVotes": 1,
            "isEditable": false,
            "groupId": "default",
            "isDeleted": false,
            "lastModified": 1617661857846
        }
    }
}
`,
			),
			ExpectedState: []byte(`

{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 1,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": null
}
`,
			),
		},
		{
			Name: "Upvote State Ahead",
			OldState: []byte(`
{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 1,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": null
}
`,
			),
			State: []byte(`
{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 3,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": {
        "title": "upVote",
        "oldCard": {
            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
            "columnId": "0",
            "message": "hello",
            "numVotes": 2,
            "isEditable": false,
            "groupId": "default",
            "isDeleted": false,
            "lastModified": 1617661857846
        },
        "newCard": {
            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
            "columnId": "0",
            "message": "hello",
            "numVotes": 3,
            "isEditable": false,
            "groupId": "default",
            "isDeleted": false,
            "lastModified": 1617661857846
        }
    }
}
`,
			),
			ExpectedState: []byte(`

{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 3,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": null
}
`,
			),
		},
		{
			Name: "Upvote State Behind",
			OldState: []byte(`
{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 5,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": null
}
`,
			),
			State: []byte(`
{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 1,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": {
        "title": "upVote",
        "oldCard": {
            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
            "columnId": "0",
            "message": "hello",
            "numVotes": 0,
            "isEditable": false,
            "groupId": "default",
            "isDeleted": false,
            "lastModified": 1617661857846
        },
        "newCard": {
            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
            "columnId": "0",
            "message": "hello",
            "numVotes": 1,
            "isEditable": false,
            "groupId": "default",
            "isDeleted": false,
            "lastModified": 1617661857846
        }
    }
}
`,
			),
			ExpectedState: []byte(`

{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 6,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": null
}
`,
			),
		},
		{
			Name: "Upvote Deleted Card",
			OldState: []byte(`
{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": true,
                            "lastModified": 1617661857846
                        }
                    ]
                },
                {
                    "id": "some-uuid",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-1",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "some-uuid",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": null
}
`,
			),
			State: []byte(`
{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 1,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": {
        "title": "upVote",
        "oldCard": {
            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
            "columnId": "0",
            "message": "hello",
            "numVotes": 0,
            "isEditable": false,
            "groupId": "default",
            "isDeleted": false,
            "lastModified": 1617661857846
        },
        "newCard": {
            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
            "columnId": "0",
            "message": "hello",
            "numVotes": 1,
            "isEditable": false,
            "groupId": "default",
            "isDeleted": false,
            "lastModified": 1617661857846
        }
    }
}
`,
			),
			ExpectedState: []byte(`

{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 1,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": true,
                            "lastModified": 1617661857846
                        }
                    ]
                },
                {
                    "id": "some-uuid",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-1",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 1,
                            "isEditable": false,
                            "groupId": "some-uuid",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": null
}
`,
			),
		},
		{
			Name: "Move Card",
			OldState: []byte(`
{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": null
}
`,
			),
			State: []byte(`
{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": true,
                            "lastModified": 1617661857846
                        }
                    ]
                },
                {
                    "id": "some-uuid",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-1",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "some-uuid",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": null
}
`,
			),
			ExpectedState: []byte(`
{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": true,
                            "lastModified": 1617661857846
                        }
                    ]
                },
                {
                    "id": "some-uuid",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-1",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "some-uuid",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": null
}
`,
			),
		},
		{
			Name: "Duplicate Cards",
			OldState: []byte(`
{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": true,
                            "lastModified": 1617661857846
                        }
                    ]
                },
                {
                    "id": "some-uuid",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-1",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "some-uuid",
                            "isDeleted": false,
                            "lastModified": 1617661857846
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": null
}
`,
			),
			State: []byte(`
{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": true,
                            "lastModified": 1617661857846
                        }
                    ]
                },
                {
                    "id": "some-other-uuid",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-1",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "some-other-uuid",
                            "isDeleted": false,
                            "lastModified": 1617661857847
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": null
}
`,
			),
			ExpectedState: []byte(`

{
    "roomId": "testroom",
    "columns": [
        {
            "id": "0",
            "title": "Good",
            "cardStyle": {
                "backgroundColor": "bg-danger"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-0",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "default",
                            "isDeleted": true,
                            "lastModified": 1617661857846
                        }
                    ]
                },
                {
                    "id": "some-uuid",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
               },
                {
                    "id": "some-other-uuid",
                    "columnId": "0",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": [
                        {
                            "id": "4a552ac9-c792-458c-bb13-0e9b300475fd-pk-1",
                            "columnId": "0",
                            "message": "hello",
                            "numVotes": 0,
                            "isEditable": false,
                            "groupId": "some-other-uuid",
                            "isDeleted": false,
                            "lastModified": 1617661857847
                        }
                    ]
                }
            ]
        },
        {
            "id": "1",
            "title": "Bad",
            "cardStyle": {
                "backgroundColor": "bg-primary"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "1",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        },
        {
            "id": "3",
            "title": "Actions",
            "cardStyle": {
                "backgroundColor": "bg-success"
            },
            "groups": [
                {
                    "id": "default",
                    "columnId": "3",
                    "isEditable": false,
                    "title": "ungrouped cards",
                    "retroCards": []
                }
            ]
        }
    ],
    "action": null
}
`,
			),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			var (
				os data.State
				st data.State
			)

			if err := json.Unmarshal(test.OldState, &os); err != nil {
				t.Fatal(err)
			}

			if err := json.Unmarshal(test.State, &st); err != nil {
				t.Fatal(err)
			}

			s := &S{}

			ms, err := s.mergeState(context.Background(), &os, &st)
			if err != nil {
				t.Fatal(err)
			}

			var es data.State
			if err := json.Unmarshal(test.ExpectedState, &es); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(&es, ms) {
				pe, err := prettify(t, es)
				if err != nil {
					t.Log(err)
					pe = es
				}

				me, err := prettify(t, ms)
				if err != nil {
					t.Log(err)
					me = ms
				}

				t.Fatalf("expected: %+v\ngot: %+v", pe, me)
			}
		})
	}
}

func prettify(t *testing.T, v interface{}) (interface{}, error) {
	t.Helper()

	prettyV, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return "", fmt.Errorf("cannot prettify %v", v)
	}

	return string(prettyV), nil
}
