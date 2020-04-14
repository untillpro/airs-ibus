# airs-ibus

## Bus object exchange protocol
Objects could be transferred over bus using IResultSender.

| data type     | over []byte chan                                             | over intf chan|
| --------------|-------------------------------------------------------------:|--------------:|
| []byte        | []byte as is                                                 | []byte as is  |   
| Map section   | {0 (BusPacketType), `sectionName`, [`path`], 0 (end marker)} | &ibus.Section |
| Map element   | {1, `name`, `jsonBytes`                                      | &ibus.Element |
| Array section | {2, `sectionName`, [`path`], 0}                              | &ibus.Section |
| Array element | {2, `jsonBytes`}                                             | &ibus.Element |
| Object section| {3, `sectonName`, [`path`], 0, 1, `jsonBytes`}               | &ibus.Section then &ibus.Element |

## Bus diagram
```dot
digraph graphname {
    node [ fontname = "Cambria" shape = "rect" fontsize = 12]
    compound=true

    ServiceTask0_33[label="Service1 for subjects 0-33"]
    ServiceTask34_66[label="Service2 for  subjects 34-66"]
    ServiceTask67_99[label="Service3 for  subjects 67-99"]

    subgraph cluster_cluster {
        label = "Cluster"
        Router
        Traefik
        ServiceTask0_33
        ServiceTask34_66
        ServiceTask67_99
        subgraph cluster_NATS{
            label = "NATS"
            Subj0_33
            Subj34_66
            Subj67_99
        }
    }
    Client

    Client -> Traefik
    Traefik -> Router
    Router -> Subj0_33
    Router -> Subj34_66
    Router -> Subj67_99
    Subj0_33 -> ServiceTask0_33
    Subj34_66 -> ServiceTask34_66
    Subj67_99 -> ServiceTask67_99
}
```