// Package app contains Echo routing, middleware, request binding, and response mapping.
//
// This layer adapts HTTP requests to application behavior and coordinates core,
// store, service, and platform packages. Domain policy should move downward into
// core or service packages when it does not require HTTP-specific state.
package app
