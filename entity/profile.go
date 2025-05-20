package entity

import (
	"fmt"
	oscmd "fwtui/utils/cmd"
	"os"
	"strings"

	"github.com/samber/lo"
)

type UFWProfile struct {
	Name        string
	Title       string
	Description string
	Ports       []string
	Installed   bool
}

func CreateProfile(p UFWProfile) string {
	content := fmt.Sprintf("[%s]\ntitle=%s\ndescription=%s\nports=%s\n",
		p.Name, p.Name, p.Title, strings.Join(p.Ports, "|"))
	err := os.WriteFile("/etc/ufw/applications.d/"+p.Name+".profile", []byte(content), 0644)
	if err != nil {
		return fmt.Sprintf("Error creating profile: %s", err)
	}
	oscmd.RunCommand(fmt.Sprintf("sudo ufw app update \"%s\"", p.Name))()
	return fmt.Sprintf("Profile %s created", p.Name)
}

func LoadInstalledProfiles() ([]UFWProfile, error) {

	out := oscmd.RunCommand("sudo ufw app list")()

	profileNames := strings.Split(strings.TrimSpace(out), "\n")[1:]

	var profiles []UFWProfile
	for _, name := range profileNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		profile, err := getUFWProfileInfo(name)
		if err != nil {
			continue
		}
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

func getUFWProfileInfo(name string) (UFWProfile, error) {
	out := oscmd.RunCommand(fmt.Sprintf("sudo ufw app info \"%s\"", name))()
	lines := strings.Split(out, "\n")

	profile := UFWProfile{
		Name:      name,
		Installed: true,
	}

	for i, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "Profile:"):
			profile.Name = strings.TrimSpace(strings.TrimPrefix(line, "Profile:"))
		case strings.HasPrefix(line, "Title:"):
			profile.Title = strings.TrimSpace(strings.TrimPrefix(line, "Title:"))
		case strings.HasPrefix(line, "Ports:") || strings.HasPrefix(line, "Port:"):
			for j := i + 1; j < len(lines); j++ {
				portLine := strings.TrimSpace(lines[j])
				if portLine == "" {
					break
				}
				profile.Ports = append(profile.Ports, portLine)
			}
		}
	}

	return profile, nil
}

func InstallableProfiles() []UFWProfile {
	installedProfiles, _ := LoadInstalledProfiles()
	installedProfileNames := lo.Map(installedProfiles, func(p UFWProfile, _ int) string {
		return p.Name
	})

	profiles := []UFWProfile{
		// Common access
		{Name: "OpenSSH", Title: "Secure shell access (SSH)", Ports: []string{"22/tcp"}, Installed: lo.Contains(installedProfileNames, "OpenSSH")},
		{Name: "HTTP", Title: "Generic HTTP service", Ports: []string{"80/tcp"}, Installed: lo.Contains(installedProfileNames, "HTTP")},
		{Name: "HTTPS", Title: "Generic HTTPS service", Ports: []string{"443/tcp"}, Installed: lo.Contains(installedProfileNames, "HTTPS")},

		// Web servers
		{Name: "Nginx HTTP", Title: "Nginx web server (HTTP only)", Ports: []string{"80/tcp"}, Installed: lo.Contains(installedProfileNames, "Nginx HTTP")},
		{Name: "Nginx HTTPS", Title: "Nginx web server (HTTPS only)", Ports: []string{"443/tcp"}, Installed: lo.Contains(installedProfileNames, "Nginx HTTPS")},
		{Name: "Nginx Full", Title: "Nginx web server (HTTP and HTTPS)", Ports: []string{"80,443/tcp"}, Installed: lo.Contains(installedProfileNames, "Nginx Full")},
		{Name: "Apache", Title: "Apache web server (HTTP only)", Ports: []string{"80/tcp"}, Installed: lo.Contains(installedProfileNames, "Apache")},
		{Name: "Apache Secure", Title: "Apache web server (HTTPS only)", Ports: []string{"443/tcp"}, Installed: lo.Contains(installedProfileNames, "Apache Secure")},
		{Name: "Apache Full", Title: "Apache web server (HTTP and HTTPS)", Ports: []string{"80,443/tcp"}, Installed: lo.Contains(installedProfileNames, "Apache Full")},

		// Databases
		{Name: "PostgreSQL", Title: "PostgreSQL database server", Ports: []string{"5432/tcp"}, Installed: lo.Contains(installedProfileNames, "PostgreSQL")},
		{Name: "MySQL", Title: "MySQL database server", Ports: []string{"3306/tcp"}, Installed: lo.Contains(installedProfileNames, "MySQL")},
		{Name: "MongoDB", Title: "MongoDB database", Ports: []string{"27017/tcp"}, Installed: lo.Contains(installedProfileNames, "MongoDB")},
		{Name: "Redis", Title: "Redis key-value store", Ports: []string{"6379/tcp"}, Installed: lo.Contains(installedProfileNames, "Redis")},
		{Name: "InfluxDB", Title: "InfluxDB time series database", Ports: []string{"8086/tcp"}, Installed: lo.Contains(installedProfileNames, "InfluxDB")},
		{Name: "Elasticsearch", Title: "Elasticsearch search engine", Ports: []string{"9200,9300/tcp"}, Installed: lo.Contains(installedProfileNames, "Elasticsearch")},

		// DevOps / containers
		{Name: "Docker Remote API", Title: "Docker remote API", Ports: []string{"2375,2376/tcp"}, Installed: lo.Contains(installedProfileNames, "Docker Remote API")},
		{Name: "Kubernetes API", Title: "Kubernetes API server", Ports: []string{"6443/tcp"}, Installed: lo.Contains(installedProfileNames, "Kubernetes API")},
		{Name: "Docker Swarm", Title: "Docker Swarm cluster communication", Ports: []string{"2377,7946/tcp", "7946,4789/udp"}, Installed: lo.Contains(installedProfileNames, "Docker Swarm")},

		// VPN
		{Name: "WireGuard", Title: "WireGuard VPN", Ports: []string{"51820/udp"}, Installed: lo.Contains(installedProfileNames, "WireGuard")},
		{Name: "OpenVPN", Title: "OpenVPN", Ports: []string{"1194/udp"}, Installed: lo.Contains(installedProfileNames, "OpenVPN")},

		// Email
		{Name: "SMTP", Title: "Simple Mail Transfer Protocol", Ports: []string{"25/tcp"}, Installed: lo.Contains(installedProfileNames, "SMTP")},
		{Name: "SMTPS", Title: "SMTP over SSL", Ports: []string{"465/tcp"}, Installed: lo.Contains(installedProfileNames, "SMTPS")},
		{Name: "Submission", Title: "Mail Submission Agent", Ports: []string{"587/tcp"}, Installed: lo.Contains(installedProfileNames, "Submission")},
		{Name: "IMAPS", Title: "IMAP over SSL", Ports: []string{"993/tcp"}, Installed: lo.Contains(installedProfileNames, "IMAPS")},
		{Name: "POP3S", Title: "POP3 over SSL", Ports: []string{"995/tcp"}, Installed: lo.Contains(installedProfileNames, "POP3S")},

		// DNS
		{Name: "DNS", Title: "Domain Name System", Ports: []string{"53/tcp", "53/udp"}, Installed: lo.Contains(installedProfileNames, "DNS")},

		// File sharing
		{Name: "Samba", Title: "Windows file/printer sharing (Samba)", Ports: []string{"137,138/udp", "139,445/tcp"}, Installed: lo.Contains(installedProfileNames, "Samba")},
		{Name: "NFS", Title: "Network File System", Ports: []string{"111,2049/tcp", "111,2049/udp"}, Installed: lo.Contains(installedProfileNames, "NFS")},

		// Misc
		{Name: "CUPS", Title: "Common Unix Printing System", Ports: []string{"631/tcp"}, Installed: lo.Contains(installedProfileNames, "CUPS")},
		{Name: "VNC", Title: "Virtual Network Computing (remote desktop)", Ports: []string{"5900/tcp"}, Installed: lo.Contains(installedProfileNames, "VNC")},
		{Name: "Deluge", Title: "Deluge BitTorrent client", Ports: []string{"6881/tcp", "6881/udp"}, Installed: lo.Contains(installedProfileNames, "Deluge")},
		{Name: "Prometheus", Title: "Prometheus monitoring", Ports: []string{"9090/tcp"}, Installed: lo.Contains(installedProfileNames, "Prometheus")},
		{Name: "Grafana", Title: "Grafana dashboards", Ports: []string{"3000/tcp"}, Installed: lo.Contains(installedProfileNames, "Grafana")},
		{Name: "RabbitMQ", Title: "RabbitMQ message broker", Ports: []string{"5672,15672/tcp"}, Installed: lo.Contains(installedProfileNames, "RabbitMQ")},
		{Name: "Mosquitto", Title: "Mosquitto MQTT broker", Ports: []string{"1883,8883/tcp"}, Installed: lo.Contains(installedProfileNames, "Mosquitto")},
	}

	profiles = lo.Filter(profiles, func(p UFWProfile, _ int) bool {
		return !p.Installed
	})

	return profiles
}
