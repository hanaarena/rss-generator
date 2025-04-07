# RSS Feed generator

```mermaid
graph TD
    A[main.go] --> B(chromedp.NewContext);
    A --> C(cacheService.NewMemoryCache);
    A --> D(providers.NewTheVergeScraper);
    A --> E(providers.NewFreeCodeCampScraper);
    A --> F(cronService.NewCronService);
    F --> G{cronService.AddTheVergeJob};
    F --> H{cronService.SetScraper};
    H --> I{cronService.AddFreeCodeCampJob};
    F --> J{cronService.Start};
    A --> K{Scraper Initialization Loop};
    K --> L{vergeScraper.Scrape};
    K --> M{freeCodeCampScraper.Scrape};
    A --> N{scraperFactories Map};
    N --> O{theverge: func};
    N --> P{freecodecamp: func};
    A --> Q["http.HandleFunc(/feed/)"];
    Q --> R{URL Path Parsing};
    R --> S{scraperFactories Lookup};
    S -- Found --> T{Factory Execution};
    T --> U{Scraper.Scrape};
    U --> V{Response Writing};
    S -- Not Found --> W{http.NotFound};
    U -- Error --> X{http.Error};
    A --> Y{http.ListenAndServe};

    style A fill:#f9f,stroke:#222,stroke-width:2px,color:#000000;
    style B fill:#ccf,stroke:#222,stroke-width:1px,color:#000000;
    style C fill:#ccf,stroke:#222,stroke-width:1px,color:#000000;
    style D fill:#ccf,stroke:#222,stroke-width:1px,color:#000000;
    style E fill:#ccf,stroke:#222,stroke-width:1px,color:#000000;
    style F fill:#ccf,stroke:#222,stroke-width:1px,color:#000000;
    style G fill:#cfc,stroke:#222,stroke-width:1px,color:#000000;
    style H fill:#cfc,stroke:#222,stroke-width:1px,color:#000000;
    style I fill:#cfc,stroke:#222,stroke-width:1px,color:#000000;
    style J fill:#cfc,stroke:#222,stroke-width:1px,color:#000000;
    style K fill:#ccf,stroke:#222,stroke-width:1px,color:#000000;
    style L fill:#cfc,stroke:#222,stroke-width:1px,color:#000000;
    style M fill:#cfc,stroke:#222,stroke-width:1px,color:#000000;
    style N fill:#ccf,stroke:#222,stroke-width:1px,color:#000000;
    style O fill:#cfc,stroke:#222,stroke-width:1px,color:#000000;
    style P fill:#cfc,stroke:#222,stroke-width:1px,color:#000000;
    style Q fill:#ccf,stroke:#222,stroke-width:1px,color:#000000;
    style R fill:#cfc,stroke:#222,stroke-width:1px,color:#000000;
    style S fill:#cfc,stroke:#222,stroke-width:1px,color:#000000;
    style T fill:#cfc,stroke:#222,stroke-width:1px,color:#000000;
    style U fill:#cfc,stroke:#222,stroke-width:1px,color:#000000;
    style V fill:#cfc,stroke:#222,stroke-width:1px,color:#000000;
    style W fill:#fcc,stroke:#222,stroke-width:1px,color:#000000;
    style X fill:#fcc,stroke:#222,stroke-width:1px,color:#000000;
    style Y fill:#ccf,stroke:#222,stroke-width:1px,color:#000000;

    classDef default fill:#ccf,stroke:#222,stroke-width:1px,color:#000000;
    classDef error fill:#fcc,stroke:#222,stroke-width:1px,color:#000000;
    classDef success fill:#cfc,stroke:#222,stroke-width:1px,color:#000000;
    classDef main fill:#f9f,stroke:#222,stroke-width:2px,color:#000000;
    class A main;
    class W,X error;
    class G,H,I,J,L,M,O,P,R,S,T,U,V success;
```